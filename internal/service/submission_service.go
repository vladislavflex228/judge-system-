package service

import (
	"context"
	"errors"
	"judge-system/internal/errs"
	"judge-system/internal/models"
	"time"
)

type CreateSubmissionDTO struct { // Структура-состояние Data - Transfer - Object
	TaskID     int64  `json:"task_id"`
	UserID     int64  `json:"user_id"`
	LanguageID int    `json:"language_id"`
	Code       string `json:"code"`
}

type SubmissionRepository interface {
	Create(ctx context.Context, sub *models.Submission) error
	GetById(ctx context.Context, id int64) (*models.Submission, error)
}

type TaskRepository interface {
	GetById(ctx context.Context, id int64) (*models.Task, error)
}

type SubmissionService interface {
	CreateSubmission(ctx context.Context, dto CreateSubmissionDTO) (*models.Submission, error)
	GetSubmissionById(ctx context.Context, id int64) (*models.Submission, error)
}

type submissionService struct { // Структура-сервис Behavior
	subRepo  SubmissionRepository
	taskRepo TaskRepository
}

func (s *submissionService) CreateSubmission(ctx context.Context, dto CreateSubmissionDTO) (*models.Submission, error) {
	if dto.UserID <= 0 || dto.TaskID <= 0 || dto.LanguageID <= 0 || dto.Code == "" {
		return nil, errs.ErrInvalidInput
	}

	if dto.LanguageID < 1 || dto.LanguageID > 3 {
		return nil, errs.ErrUnsupportedLanguage
	}

	if len(dto.Code) > 65536 {
		return nil, errs.ErrCodeTooLarge
	}

	if _, err := s.taskRepo.GetById(ctx, dto.TaskID); err != nil {
		if errors.Is(err, errs.ErrTaskNotFound) {
			return nil, errs.ErrNotFound
		}

		return nil, err
	}

	sub := models.NewSubmission(dto.TaskID, dto.UserID, dto.LanguageID, 0, 0, dto.Code, "pending", time.Now())

	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *submissionService) GetSubmissionById(ctx context.Context, id int64) (*models.Submission, error) {

	submission, err := s.subRepo.GetById(ctx, id)

	if err != nil {
		if errors.Is(err, errs.ErrSubmissionNotFound) {
			return nil, errs.ErrNotFound
		}

		return nil, err
	}

	return submission, nil
}

func NewSubmissionService(subRepo SubmissionRepository, taskRepo TaskRepository) SubmissionService { // Polymorphism
	return &submissionService{
		subRepo:  subRepo,
		taskRepo: taskRepo,
	}
}
