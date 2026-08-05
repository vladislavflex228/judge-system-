package middleware

import (
	"context"
	"fmt"
	"judge-system/internal/auth"
	"judge-system/internal/responces"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

type contextKey string

const UserIDKey contextKey = "userID"
const UserRoleKey contextKey = "userRole"

// Authorization: Bearer <ваш_токен>
func AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			responces.ResponseError(w, http.StatusUnauthorized, fmt.Errorf("Unathorized").Error(), "Empty AuthHeader")
			return
		}

		authContent := strings.SplitN(authHeader, " ", 2)

		if len(authContent) != 2 || authContent[0] != "Bearer" {
			responces.ResponseError(w, http.StatusUnauthorized, fmt.Errorf("Invalid token format").Error(), "Format:Bearer <your_token>")
			return
		}

		jwtKey := os.Getenv("JWT_KEY")

		claims := &auth.CustomClaims{}

		token, err := jwt.ParseWithClaims(authContent[1], claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok { //type assertion
				return nil, jwt.ErrSignatureInvalid
			}

			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			responces.ResponseError(w, http.StatusUnauthorized, fmt.Errorf("Invalid or expired token").Error(), "")
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
