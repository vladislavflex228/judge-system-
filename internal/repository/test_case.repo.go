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
	test := &models.TestCase{}

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

func (t *TestRepository) GetAllTestsByTaskID(ctx context.Context, task_id int64) ([]models.TestCase, error) {
	query := `SELECT id,task_id,title,description,input_data,output_data,is_hidden,created_at
			  FROM test_cases
			  WHERE task_id = $1
			  ORDER BY id`

	row_cursor, err := t.db.Query(ctx, query, task_id)

	if err != nil {
		return nil, fmt.Errorf("query error at func GetAllTestsByTaskID : %w", err)
	}

	defer row_cursor.Close() //row_cursor - сетевой курсор , подключенный к базе данных , требует отключения

	tests := []models.TestCase{}

	for row_cursor.Next() {
		test := models.TestCase{}

		err := row_cursor.Scan(
			&test.ID,
			&test.TaskID,
			&test.Title,
			&test.Description,
			&test.InputData,
			&test.OutputData,
			&test.IsHidden)

		if err != nil {
			return nil, fmt.Errorf("scan error at func GetAllTestsByTaskID : %w ", err)
		}

		tests = append(tests, test)
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
		return nil, fmt.Errorf("error at Getalltestsbyidslice : %w", err)
	}

	defer row_cursor.Close()

	testCases := []models.TestCase{}

	for row_cursor.Next() {
		test := &models.TestCase{}

		err := row_cursor.Scan(
			&test.ID,
			&test.TaskID,
			&test.Title,
			&test.Description,
			&test.InputData,
			&test.OutputData,
			&test.IsHidden)

		if err != nil {
			return nil, fmt.Errorf("scan error at Getalltestsbyidslice : %w", err)
		}

		testCases = append(testCases, *test)
	}

	return testCases, nil

}
