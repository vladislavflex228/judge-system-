package service

import (
	"context"
	"judge-system/internal/auth"
	"judge-system/internal/models"
)

type RegUserRepository interface {
	Create(ctx context.Context, user *models.User) error
}

type RegService interface {
	Produce(ctx context.Context, username, email, password string) error
}

type regService struct {
	repo RegUserRepository
}

func (r *regService) Produce(ctx context.Context, username, email, password string) error {
	hash_password, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	user := &models.User{Username: username, Email: email, HashPassword: hash_password}

	err = r.repo.Create(ctx, user)

	return err
}

func NewRegService(userRepo RegUserRepository) RegService {
	return &regService{repo: userRepo}
}
