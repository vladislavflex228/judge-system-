package service

import (
	"context"
	"errors"
	"judge-system/internal/errs"
)

type LoginUserRepository interface {
	GetCredentialsByEmail(ctx context.Context, email string) (int64, string, error)
}

type LoginService interface {
	GetCredentialsByEmail(ctx context.Context, email string) (int64, string, error)
}

type loginService struct {
	repo LoginUserRepository
}

func (s *loginService) GetCredentialsByEmail(ctx context.Context, email string) (int64, string, error) {
	id, hash_password, err := s.repo.GetCredentialsByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return 0, "", errs.ErrNotFound

		}
		return 0, "", err
	}

	return id, hash_password, nil

}

func NewLoginService(repo LoginUserRepository) LoginService {
	return &loginService{repo: repo}
}
