package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"judge-system/internal/models"
)

type TestRepository struct {
	db *pgxpool.Pool
}

func NewTestRepository(db *pgxpool.Pool) *TestRepository {
	return &TestRepository{db: db}
}

func (t *TestRepository) Create(ctx context.Context, test *models.TestCase) error {
	query := `INSERT INTO test_cases (task_id,title,description,input_data,output_data,is_hidden,created_at)
			  VALUES ($1,$2,$3,$4,$5,$6,NOW())
			  RETURNING id,created_at`

	err := t.db.QueryRow(ctx, query, test.TaskID, test.Title, test.Description, test.InputData, test.OutputData, test.IsHidden).Scan(&test.ID, &test.CreatedAt)

	if err != nil {
		return fmt.Errorf("create test_case_repo error : %w", err)
	}

	return nil

}

func (t *TestRepository) GetById(ctx context.Context, id int64) (*models.TestCase, error) {
	query := `SELECT id,task_id,title,description,input_data,output_data,is_hidden,created_at
			  FROM test_cases
			  WHERE id = $1`
	var test *models.TestCase

	err := t.db.QueryRow(ctx, query, id).Scan(
		&test.ID,
		&test.TaskID,
		&test.Title,
		&test.Description,
		&test.InputData,
		&test.OutputData,
		&test.IsHidden)

	if err != nil {
		return nil, fmt.Errorf("GetById test_case_repo error: %w", err)
	}

	return test, nil
}
