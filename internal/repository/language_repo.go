package repository

import (
	"context"
	"errors"
	"fmt"
	"judge-system/internal/errs"
	"judge-system/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LanguageRepository struct {
	db *pgxpool.Pool
}

func NewLanguageRepository(db *pgxpool.Pool) *LanguageRepository {
	return &LanguageRepository{db: db}
}

func (l *LanguageRepository) Create(ctx context.Context, language *models.Language) error {
	query := `INSERT INTO languages (name,slug,build_command,execute_command,is_active)
			  VALUES ($1,$2,$3,$4,$5,)
			  RETURNING id`
	err := l.db.QueryRow(
		ctx, query,
		language.Name,
		language.Slug,
		language.BuildCmd,
		language.ExeCmd,
		language.IsActive).Scan(&language.ID)

	if err != nil {
		return fmt.Errorf("language repository : create : %w", err)
	}

	return nil
}

func (l *LanguageRepository) GetById(ctx context.Context, id int) (*models.Language, error) {
	query := `SELECT id,name,slug,build_command,execute_command,is_active
			  FROM languages
			  WHERE id = $1`

	var language models.Language

	err := l.db.QueryRow(ctx, query, id).Scan(
		&language.ID,
		&language.Name,
		&language.Slug,
		&language.BuildCmd,
		&language.ExeCmd,
		&language.IsActive)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrLanguageNotFound
		}
		return nil, fmt.Errorf("language repository : get by id : %w", err)
	}

	return &language, nil
}
