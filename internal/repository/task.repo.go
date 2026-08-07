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

type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

func (t *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	query := `INSERT INTO tasks (title,description,time_limit_ms,memory_limit_mb,created_at)
			  VALUES ($1,$2,$3,$4,NOW())
			  RETURNING id, created_at`
	err := t.db.QueryRow(
		ctx, query,
		task.Title,
		task.Description,
		task.TimeLimit,
		task.MemoryLimit).Scan(&task.ID, &task.CreatedAt)

	if err != nil {
		return fmt.Errorf("task repository : create : %w", err)
	}

	return nil
}

func (t *TaskRepository) GetById(ctx context.Context, id int64) (*models.Task, error) {
	query := `SELECT id,title,description,time_limit_ms,memory_limit_mb,created_at
			  FROM tasks
			  WHERE id = $1`

	var task models.Task

	err := t.db.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.TimeLimit,
		&task.MemoryLimit,
		&task.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrTaskNotFound
		}
		return nil, fmt.Errorf("task repository : get by id : %w", err)
	}

	return &task, nil
}
