package service

import (
	"context"
	"errors"
	"judge-system/internal/models"
	"time"
)

var (
	ErrInvalidInput        = errors.New("invalid input parameters")
	ErrUnsupportedLanguage = errors.New("unsupported programming language")
	ErrCodeTooLarge        = errors.New("code size exceeds maximum limit")
	ErrTaskNotFound        = errors.New("task not found")
	ErrSubmissionNotFound  = errors.New("submission not found")
)

type CreateSubmissionDTO struct { // Структура состояние Data - Transfer - Object
	TaskID     int64
	UserID     int64
	LanguageID int
	Code       string
}

type SubmissionRepository interface {
	Create(ctx context.Context, sub *models.Submission) error
	GetByID(ctx context.Context, id int64) (*models.Submission, error)
}

type TaskRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Task, error)
}

type SubmissionService interface {
	CreateSubmission(ctx context.Context, dto CreateSubmissionDTO) (*models.Submission, error)
	GetSubmissionByID(ctx context.Context, id int64) (*models.Submission, error)
}

type submissionService struct { // Структура-сервис Behavior
	subRepo  SubmissionRepository
	taskRepo TaskRepository
}

func (s *submissionService) CreateSubmission(ctx context.Context, dto CreateSubmissionDTO) (*models.Submission, error) {
	if dto.UserID <= 0 || dto.TaskID <= 0 || dto.LanguageID <= 0 || dto.Code == "" {
		return nil, ErrInvalidInput
	}

	if dto.LanguageID < 1 || dto.LanguageID > 3 {
		return nil, ErrUnsupportedLanguage
	}

	if len(dto.Code) > 65536 {
		return nil, ErrCodeTooLarge
	}

	if _, err := s.taskRepo.GetByID(ctx, dto.TaskID); err != nil {
		return nil, ErrTaskNotFound
	}

	sub := models.NewSubmission(dto.TaskID, dto.UserID, dto.LanguageID, 0, 0, dto.Code, "pending", time.Now())

	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *submissionService) GetSubmissionByID(ctx context.Context, id int64) (*models.Submission, error) {

	submission, err := s.subRepo.GetByID(ctx, id)

	if err != nil {
		return nil, err
	}

	return submission, nil
}

func NewSubmissionService(subRepo SubmissionRepository, taskRepo TaskRepository) SubmissionService {
	return &submissionService{
		subRepo:  subRepo,
		taskRepo: taskRepo,
	}
}
