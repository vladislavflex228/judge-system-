package repository

import (
	"context"
	"fmt"
	"judge-system/internal/models"

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
		return fmt.Errorf("create language_repo error: %w", err)
	}

	return nil
}

func (l *LanguageRepository) GetById(ctx context.Context, id int64) (*models.Language, error) {
	query := `SELECT id,name,slug,build_command,execute_command,is_active
			  FROM languages
			  WHERE id = $1`

	language := &models.Language{}

	err := l.db.QueryRow(ctx, query, id).Scan(
		&language.ID,
		&language.Name,
		&language.Slug,
		&language.BuildCmd,
		&language.ExeCmd,
		&language.IsActive)

	if err != nil {
		return nil, fmt.Errorf("getbyid language_repo error : %w", err)
	}

	return language, nil
}
