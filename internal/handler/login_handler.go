package handler

import (
	"context"
	"encoding/json"
	"errors"
	"judge-system/internal/auth"
	"judge-system/internal/responces"
	"judge-system/internal/service"
	"log/slog"
	"net/http"
	"time"
)

type LoginHandler struct {
	svr service.LoginService
}

func NewLoginHandler(svr service.LoginService) *LoginHandler {
	return &LoginHandler{svr: svr}
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginOutput struct {
	Status string `json:"status"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

func (l *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	var in LoginInput
	err := json.NewDecoder(r.Body).Decode(&in)

	if err != nil {
		slog.Warn("Login failed:bad JSON syntax", slog.Any("err", err))
		responces.ResponseError(w, http.StatusBadRequest, "Invalid JSON format", "Expected email,password fields")
		return
	}

	if in.Email == "" || in.Password == "" {
		slog.Warn("Login failed:password or email are empty", slog.String("reason", "empty strings"))
		responces.ResponseError(w, http.StatusBadRequest, "Missing credentials", "Email and password must be non-empty")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)

	defer cancel()

	id, hash_password, err := l.svr.GetCredentialsByEmail(ctx, in.Email)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			slog.Warn("Login failed:email not found", slog.String("email", in.Email))
			responces.ResponseError(w, http.StatusUnauthorized, "Invalid credentials", "Wrong password or credentials") //Защита от хакера : вернуть статус 401 вместо 404
			return
		default:
			slog.Error("Database error during login", slog.String("email", in.Email), slog.Any("err", err))
			responces.ResponseError(w, http.StatusInternalServerError, "Internal server error", "Try again later")
			return
		}
	}

	if ok := auth.CheckPasswordHash(in.Password, hash_password); !ok {
		slog.Warn("Login failed: wrong password", slog.String("email", in.Email))
		responces.ResponseError(w, http.StatusUnauthorized, "Invalid credentials", "Wrong password or email")
		return
	}

	token, err := auth.GenerateToken(id)
	if err != nil {
		slog.Error("Failed to generate JWT token", slog.Int64("user_id", id), slog.Any("err", err))
		responces.ResponseError(w, http.StatusInternalServerError, "Internal error", "Generating token error")
	}

	slog.Info("User logged in successfully", slog.Int64("user_id", id), slog.String("email", in.Email))

	responces.ResponseJson(w, http.StatusAccepted, LoginOutput{Status: "OK", Email: in.Email, Token: token})
}

//slog.Info (Важные вехи): Запуск сервера, успешный старт миграций, успешный вход пользователя
//slog.Warn (Ошибки клиента): Неверный пароль, просроченный токен, плохой JSON в body.
// Сервер работает штатно, но клиент делает что-то не то.
//
//slog.Error (Критические сбои бэкенда): Упал SQL-запрос, отвалилась сеть с базой данных,
//сработал recover после паники
