package handler

import (
	"context"
	"encoding/json"
	"errors"
	"judge-system/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

func ResponseJson(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

type ErrorBody struct {
	Error struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	} `json:"error"`
}

func ResponseError(w http.ResponseWriter, status int, code, message string) {
	var resp ErrorBody

	resp.Error.Code = code
	resp.Error.Message = message

	ResponseJson(w, status, resp)
}

type SubmissionHandler struct {
	svr service.SubmissionService
}

func NewSubmissionHandler(svr service.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{svr: svr}
}

func (h *SubmissionHandler) Create(w http.ResponseWriter, r *http.Request) {

	var dto service.CreateSubmissionDTO

	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		ResponseError(w, http.StatusBadRequest, err.Error(), "Bad request")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)

	defer cancel()

	sub, subErr := h.svr.CreateSubmission(ctx, dto)

	if subErr != nil {
		switch {
		case errors.Is(subErr, context.DeadlineExceeded):
			ResponseError(w, http.StatusGatewayTimeout, subErr.Error(), "Deadline of submission_service was exceeded")

		default:
			ResponseError(w, http.StatusInternalServerError, subErr.Error(), "Internal error")
		}

		return
	}

	ResponseJson(w, http.StatusAccepted, sub)
}

func (h *SubmissionHandler) Get(w http.ResponseWriter, r *http.Request) {
	strId := r.URL.Query().Get("id")

	id, err := strconv.ParseInt(strId, 10, 64)

	if err != nil {
		ResponseError(w, http.StatusBadRequest, err.Error(), "Bad request")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	sub, subErr := h.svr.GetSubmissionByID(ctx, id)

	if subErr != nil {
		switch {
		case errors.Is(subErr, context.DeadlineExceeded):
			ResponseError(w, http.StatusGatewayTimeout, subErr.Error(), "Deadline of submission_service was exceeded")
		case errors.Is(subErr, pgx.ErrNoRows):
			ResponseError(w, http.StatusNotFound, subErr.Error(), "Submission not found")
		default:
			ResponseError(w, http.StatusInternalServerError, subErr.Error(), "Internal error")
		}

		return

	}

	ResponseJson(w, http.StatusAccepted, sub)

}
