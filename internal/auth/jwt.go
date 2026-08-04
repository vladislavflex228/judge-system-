package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	user_id int64
	role    string
	jwt.RegisteredClaims
}

func GenerateToken(user_id int64, role string) (string, error) {
	claims := &CustomClaims{
		user_id: user_id,
		role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtKey := os.Getenv("JWT_SECRET")

	return token.SignedString(jwtKey)
}
