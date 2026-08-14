package handler

import (
	"context"
	"errors"
	"judge-system/internal/errs"
	"judge-system/internal/responces"
	"judge-system/internal/service"
	"log/slog"
	"net/http"
	"strconv"
)

type TaskHandler struct {
	svr service.JudgeService
}

func NewTaskHandler(svr service.JudgeService) *TaskHandler {
	return &TaskHandler{svr: svr}
}

type JudgeOutput struct {
	Status string `json:"status"`
	Id     int64  `json:"id"`
}

func (h *TaskHandler) Judge(w http.ResponseWriter, r *http.Request) {
	strId := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(strId, 10, 64)

	if err != nil {
		slog.Warn("Judge failed : bad JSON format", slog.Any("err", err))
		responces.ResponseError(w, http.StatusBadRequest, "Invalid JSON format", "Expected id field")
		return
	}

	//	ctxTimeout, cancel := context.WithTimeout(r.Context(), 15*time.Second)

	//	defer cancel()

	status, judgeErr := h.svr.JudgeSubmission(r.Context(), id)

	if judgeErr != nil {
		switch {
		case errors.Is(judgeErr, context.DeadlineExceeded):
			slog.Error("Judge failed : deadline of responce from bd", slog.Any("err", judgeErr))
			responces.ResponseError(w, http.StatusGatewayTimeout, "Internal server problem", "Try again later")
		case errors.Is(judgeErr, errs.ErrNotFound):
			slog.Error("Judge failed : sub not found", slog.Any("err", judgeErr))
			responces.ResponseError(w, http.StatusNotFound, "Non-existed submission id", "Write correct id")
		default:
			slog.Error("Judge failed : bd error", slog.Any("err", judgeErr))
			responces.ResponseError(w, http.StatusInternalServerError, "Internal server problem", "Try again later")
		}

		return
	}

	slog.Info("Submission was judged succesfully", slog.Int64("id", id))
	responces.ResponseJson(w, http.StatusAccepted, JudgeOutput{Status: status, Id: id})
}
