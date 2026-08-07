package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"judge-system/internal/errs"
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
		return fmt.Errorf("user repository : create : %w", err)
	}

	return nil

}

func (u *UserRepository) GetById(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id,username,email,hash_password,created_at
			  FROM users
			  WHERE id = $1`
	var user models.User

	err := u.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.HashPassword,
		&user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}
		return nil, fmt.Errorf("user repository : get by id : %w", err)
	}

	return &user, nil
}

func (u *UserRepository) GetCredentialsByEmail(ctx context.Context, email string) (int64, string, error) {
	query := `
	SELECT id,hash_password
	FROM users
	WHERE email = $1`

	var id int64
	var hash_password string
	err := u.db.QueryRow(ctx, query, email).Scan(&id, &hash_password)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", errs.ErrUserNotFound
		}
		return 0, "", fmt.Errorf("user repository : get credentials by id : %w", err)
	}

	return id, hash_password, nil
}
