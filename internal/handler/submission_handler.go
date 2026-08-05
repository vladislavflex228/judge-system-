package handler

import (
	"context"
	"encoding/json"
	"errors"
	"judge-system/internal/responces"
	"judge-system/internal/service"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type SubmissionHandler struct {
	svr service.SubmissionService
}

func NewSubmissionHandler(svr service.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{svr: svr}
}

func (h *SubmissionHandler) Create(w http.ResponseWriter, r *http.Request) {

	var dto service.CreateSubmissionDTO

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		slog.Warn("Create submission failed : bad JSON format", slog.Any("err", err))
		responces.ResponseError(w, http.StatusBadRequest, "Invalid JSON format", "Expected task_id,user_id,language_id,code fields")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)

	defer cancel()

	sub, subErr := h.svr.CreateSubmission(ctx, dto)

	if subErr != nil {
		switch {
		case errors.Is(subErr, context.DeadlineExceeded):
			slog.Error("Create submission failed: deadline od responce from bd exceeded", slog.Any("err", subErr))
			responces.ResponseError(w, http.StatusGatewayTimeout, "Internal server error", "Try again later")

		default:
			slog.Error("Create submission failed : bd error", slog.Any("err", subErr))
			responces.ResponseError(w, http.StatusInternalServerError, subErr.Error(), "Internal error")
		}

		return
	}

	slog.Info("Submission was created", slog.Int64("id", sub.ID))
	responces.ResponseJson(w, http.StatusAccepted, sub)
}

func (h *SubmissionHandler) Get(w http.ResponseWriter, r *http.Request) {
	strId := r.URL.Query().Get("id")

	id, err := strconv.ParseInt(strId, 10, 64)

	if err != nil {
		slog.Warn("Create submission failed : bad JSON format", slog.Any("err", err))
		responces.ResponseError(w, http.StatusBadRequest, "Invalid JSON format", "Expected id field")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	sub, subErr := h.svr.GetSubmissionByID(ctx, id)

	if subErr != nil {
		switch {
		case errors.Is(subErr, context.DeadlineExceeded):
			slog.Error("Get submission failed : deadline of responce from bd", slog.Any("err", subErr))
			responces.ResponseError(w, http.StatusGatewayTimeout, "Internal server problem", "Try again later")
		case errors.Is(subErr, service.ErrSubmissionNotFound):
			slog.Warn("Get submission failed : submission not found", slog.Any("err", subErr))
			responces.ResponseError(w, http.StatusNotFound, "Non-existed submission id", "Write correct id")
		default:
			slog.Error("Get submission failed : bd error", slog.Any("err", subErr))
			responces.ResponseError(w, http.StatusInternalServerError, "Internal server problem", "Try again later")
		}

		return

	}

	slog.Info("Submission was acquiered", slog.Int64("id", id))
	responces.ResponseJson(w, http.StatusAccepted, sub)

}
