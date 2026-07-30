package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"judge-system/internal/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (username,email,hash_password,created_at)
			  VALUES ($1,$2,$3,NOW())
			  RETURNING id,created_at`

	err := u.db.QueryRow(ctx, query, user.Username, user.Email, user.HashPassword).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return fmt.Errorf("create user_repo error : %w", err)
	}

	return nil

}

func (u *UserRepository) GetById(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id,username,email,hash_password,created_at
			  FROM users
			  WHERE id = $1`
	var user *models.User

	err := u.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.HashPassword,
		&user.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("UserById user_repo error: %w", err)
	}

	return user, nil
}
