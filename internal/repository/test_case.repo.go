package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"judge-system/internal/errs"
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
		return fmt.Errorf("test_case repository : create : %w", err)
	}

	return nil

}

func (t *TestRepository) GetById(ctx context.Context, id int64) (*models.TestCase, error) {
	query := `SELECT id,task_id,title,description,input_data,output_data,is_hidden,created_at
			  FROM test_cases
			  WHERE id = $1`
	var test models.TestCase

	err := t.db.QueryRow(ctx, query, id).Scan(
		&test.ID,
		&test.TaskID,
		&test.Title,
		&test.Description,
		&test.InputData,
		&test.OutputData,
		&test.IsHidden)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrTestNotFound
		}
		return nil, fmt.Errorf("test_case repository : get by id : %w", err)
	}
	return &test, nil
}

func (t *TestRepository) GetAllTestsByTaskID(ctx context.Context, task_id int64) ([]models.TestCase, error) {
	query := `SELECT id,task_id,title,description,input_data,output_data,is_hidden,created_at
			  FROM test_cases
			  WHERE task_id = $1
			  ORDER BY id`

	row_cursor, err := t.db.Query(ctx, query, task_id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrTestNotFound
		}
		return nil, fmt.Errorf("test_case repository : get all tests by task id : %w", err)
	}

	defer row_cursor.Close() //row_cursor - сетевой курсор , подключенный к базе данных , требует отключения

	tests := make([]models.TestCase, 0)

	for row_cursor.Next() {
		var test models.TestCase

		err := row_cursor.Scan(
			&test.ID,
			&test.TaskID,
			&test.Title,
			&test.Description,
			&test.InputData,
			&test.OutputData,
			&test.IsHidden,
			&test.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("test_case repository : get all tests by task id : %w", err)
		}

		tests = append(tests, test)
	}

	if err := row_cursor.Err(); err != nil {
		return nil, fmt.Errorf("test_case repository : get all tests by task id (cursor): %w", err)
	}

	return tests, nil
}

func (t *TestRepository) GetAllTestsByIdSlice(ctx context.Context, testsId []int64) ([]models.TestCase, error) {
	query := `
	SELECT t.id,t.task_id,t.title,t.description,t.input_data,t.output_data,t.is_hidden,t.created_at
	FROM test_cases t
		INNER JOIN UNNEST($1::int8[]) AS tests_id(id)
			ON tests_id.id = t.id`

	row_cursor, err := t.db.Query(ctx, query, testsId)

	if err != nil {
		return nil, fmt.Errorf("test_case repository : get all tests by id slice : %w", err)
	}

	slog.Info("Тут")

	defer row_cursor.Close()

	testCases := make([]models.TestCase, 0)

	for row_cursor.Next() {
		var test models.TestCase

		err := row_cursor.Scan(
			&test.ID,
			&test.TaskID,
			&test.Title,
			&test.Description,
			&test.InputData,
			&test.OutputData,
			&test.IsHidden,
			&test.CreatedAt)

		if err != nil {
			return nil, fmt.Errorf("test_case repository : get all tests by id slice : %w", err)
		}

		testCases = append(testCases, test)
	}

	if err := row_cursor.Err(); err != nil {
		return nil, fmt.Errorf("test_case repository : get all tests by id slice(cursor) : %w", err)
	}

	return testCases, nil

}
