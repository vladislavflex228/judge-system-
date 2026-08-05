package handler

import (
	"context"
	"encoding/json"
	"judge-system/internal/responces"
	"judge-system/internal/service"
	"log/slog"
	"net/http"
	"time"
)

type RegHandler struct {
	svr service.RegService
}

func NewRegHandler(svr service.RegService) *RegHandler {
	return &RegHandler{svr: svr}
}

type RegInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegOutput struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Username string `json:"username"`
}

func (h *RegHandler) Registration(w http.ResponseWriter, r *http.Request) {
	in := RegInput{}
	err := json.NewDecoder(r.Body).Decode(&in)

	if err != nil {
		slog.Warn("Reg failed: bad lson syntax", slog.Any("err", err))
		responces.ResponseError(w, http.StatusBadRequest, "Invalid JSON format", "Expected username,email,password fields")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.svr.Produce(ctx, in.Username, in.Email, in.Password); err != nil {
		slog.Error("Database error during registration", slog.Any("err", err))
		responces.ResponseError(w, http.StatusInternalServerError, "Internal server error", "Try again later")
		return
	}

	slog.Info("User registered succesfully", slog.String("username", in.Username), slog.String("email", in.Email))
	responces.ResponseJson(w, http.StatusAccepted, RegOutput{Status: "OK", Username: in.Username, Message: "Registration passed succesfully"})
}
