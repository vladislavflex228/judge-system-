package repository

import (
	"context"
	"errors"
	"fmt"
	"judge-system/internal/models"
	"judge-system/internal/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubmissionRepository struct {
	db *pgxpool.Pool
}

func NewSubmissionRepository(db *pgxpool.Pool) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (s *SubmissionRepository) Create(ctx context.Context, submission *models.Submission) error {
	query := `INSERT INTO submissions (task_id,user_id,language_id,code,status,execution_time_ms,memory_used_mb,created_at)
			  VALUES ($1,$2,$3,$4,$5,$6,$7,NOW())
			  RETURNING id , created_at`
	err := s.db.QueryRow(
		ctx, query,
		submission.TaskID,
		submission.UserID,
		submission.LanguageID,
		submission.Code,
		submission.Status,
		submission.ExecutionTime,
		submission.MemoryUsed).Scan(&submission.ID, &submission.CreatedAt)

	if err != nil {
		return fmt.Errorf("create submission_repo error: %w", err)
	}

	return nil
}

func (s *SubmissionRepository) GetById(ctx context.Context, id int64) (*models.Submission, error) {
	query := `SELECT id,task_id,user_id,language_id,code,status,execution_time_ms,memory_used_mb,created_at
			  FROM submissions
			  WHERE id = $1`

	submission := &models.Submission{}

	err := s.db.QueryRow(ctx, query, id).Scan(
		&submission.ID,
		&submission.TaskID,
		&submission.UserID,
		&submission.LanguageID,
		&submission.Code,
		&submission.Status,
		&submission.ExecutionTime,
		&submission.MemoryUsed)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrSubmissionNotFound
		}
		return nil, fmt.Errorf("getbyid submission_repo error : %w", err)
	}

	return submission, nil
}

func (s *SubmissionRepository) GetAllSubmissionsByTaskID(ctx context.Context, task_id int64) ([]models.Submission, error) {
	query := `SELECT id,task_id,user_id,language_id,code,status,execution_time_ms,memory_used_mb,created_at
			  FROM submissions
			  WHERE task_id = $1
			  ORDER BY id`

	row_cursor, err := s.db.Query(ctx, query, task_id)

	if err != nil {
		return nil, fmt.Errorf("query error at func GetAllSubmissionsByTaskID : %w", err)
	}

	defer row_cursor.Close() //row_cursor - сетевой курсор , подключенный к базе данных , требует отключения

	submissions := []models.Submission{}

	for row_cursor.Next() {
		submission := models.Submission{}

		err := row_cursor.Scan(
			&submission.ID,
			&submission.TaskID,
			&submission.UserID,
			&submission.LanguageID,
			&submission.Code,
			&submission.Status,
			&submission.ExecutionTime,
			&submission.MemoryUsed)

		if err != nil {
			return nil, fmt.Errorf("scan error at func GetAllSubmissionsByTaskID : %w ", err)
		}

		submissions = append(submissions, submission)
	}

	return submissions, nil
}

func (s *SubmissionRepository) GetAllSubmissionsByUserID(ctx context.Context, user_id int64) ([]models.Submission, error) {
	query := `SELECT id,task_id,user_id,language_id,code,status,execution_time_ms,memory_used_mb,created_at
			  FROM submissions
			  WHERE user_id = $1
			  ORDER BY id`

	row_cursor, err := s.db.Query(ctx, query, user_id)

	if err != nil {
		return nil, fmt.Errorf("query error at func GetAllSubmissionsByUserID : %w", err)
	}

	defer row_cursor.Close() //row_cursor - сетевой курсор , подключенный к базе данных , требует отключения

	submissions := []models.Submission{}

	for row_cursor.Next() {
		submission := models.Submission{}

		err := row_cursor.Scan(
			&submission.ID,
			&submission.TaskID,
			&submission.UserID,
			&submission.LanguageID,
			&submission.Code,
			&submission.Status,
			&submission.ExecutionTime,
			&submission.MemoryUsed)

		if err != nil {
			return nil, fmt.Errorf("scan error at func GetAllSubmissionsByUserID : %w ", err)
		}

		submissions = append(submissions, submission)
	}

	return submissions, nil
}

func (s *SubmissionRepository) GetAllTestsIdForSubmission(ctx context.Context, id int64) ([]int64, error) {
	query := `
	SELECT t.id
	FROM submissions s
		INNER JOIN test_cases t
			ON s.task_id = t.task_id
	WHERE s.id = $1`

	row_cursor, err := s.db.Query(ctx, query, id)

	if err != nil {
		return nil, fmt.Errorf("error at func Getalltestsbysub : %w", err)
	}

	defer row_cursor.Close()

	testCasesId := []int64{}

	for row_cursor.Next() {
		var test_id int64
		err := row_cursor.Scan(&test_id)
		if err != nil {
			return nil, err
		}
		testCasesId = append(testCasesId, test_id)
	}

	return testCasesId, nil
}

func (s *SubmissionRepository) SaveById(ctx context.Context, id int64, finalVerdict string, maxTime, maxMem int) error {
	query := `
	UPDATE submissions 
	SET status = $1,execution_time_ms = $2, memory_used_mb = $3
	WHERE id = $4;`

	_, err := s.db.Exec(ctx, query, finalVerdict, maxTime, maxMem, id) // Если мы не делаем SELECT и RETURNING , используем Exec

	if err != nil {
		return err
	}

	return nil
}
