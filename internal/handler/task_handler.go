package handler

import (
	"context"
	"errors"
	"judge-system/internal/service"
	"net/http"
	"strconv"
	"time"
)

type TaskHandler struct {
	svr service.JudgeService
}

func NewTaskHandler(svr service.JudgeService) *TaskHandler {
	return &TaskHandler{svr: svr}
}

func (h *TaskHandler) Judge(w http.ResponseWriter, r *http.Request) {
	strId := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(strId, 10, 64)

	ResponseError(w, http.StatusBadRequest, err.Error(), "Judge error bad request")

	ctxTimeout, cancel := context.WithTimeout(r.Context(), 3*time.Second)

	defer cancel()

	judgeErr := h.svr.JudgeSubmission(ctxTimeout, id)

	if judgeErr != nil {
		switch {
		case errors.Is(judgeErr, context.DeadlineExceeded):
			ResponseError(w, http.StatusGatewayTimeout, judgeErr.Error(), "DeadlineExceeded error at Judge")
		default:
			ResponseError(w, http.StatusInternalServerError, judgeErr.Error(), "Internal error at Judge")
		}
	}

	ResponseJson(w, http.StatusAccepted, "OK")
}
