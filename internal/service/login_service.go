package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("resource not found")

type LoginUserRepository interface {
	CredentialsByEmail(ctx context.Context, email string) (int64, string, error)
}

type LoginService interface {
	GetCredentialsByEmail(ctx context.Context, email string) (int64, string, error)
}

type loginService struct {
	repo LoginUserRepository
}

func (s *loginService) GetPassword(ctx context.Context, email string) (int64, string, error) {
	if email == "" {
		return 0, "", fmt.Errorf("Empty email")
	}

	id, hash_password, err := s.repo.CredentialsByEmail(ctx, email)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return 0, "", ErrNotFound
		default:
			return 0, "", err

		}
	}

	return id, hash_password, nil

}
