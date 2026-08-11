package service

import (
	"context"
	"errors"
	"fmt"
	"judge-system/internal/errs"
	"judge-system/internal/judge/config"
	"judge-system/internal/judge/domain"
	"judge-system/internal/models"
	"log/slog"
	"strings"
)

type JudgeSubmissionRepository interface {
	GetAllTestsIdForSubmission(ctx context.Context, id int64) ([]int64, error)
	GetById(ctx context.Context, id int64) (*models.Submission, error)
	SaveById(ctx context.Context, id int64, finalVerdict string, maxTime, maxMem int) error
}

type JudgeLanguageRepository interface {
	GetById(ctx context.Context, id int) (*models.Language, error)
}

type JudgeTaskRepository interface {
	GetById(ctx context.Context, id int64) (*models.Task, error)
}

type JudgeTestRepository interface {
	GetAllTestsByIdSlice(ctx context.Context, testsId []int64) ([]models.TestCase, error)
}

type JudgeService interface {
	JudgeSubmission(ctx context.Context, id int64) error
}

type judgeService struct {
	subRepo  JudgeSubmissionRepository
	langRepo JudgeLanguageRepository
	taskRepo JudgeTaskRepository
	testRepo JudgeTestRepository
	runner   domain.CodeRunner
}

func (s *judgeService) JudgeSubmission(ctx context.Context, id int64) error {

	if id < 0 {
		return errs.ErrInvalidInput
	}

	sub, err := s.subRepo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrSubmissionNotFound) {
			return errs.ErrNotFound
		}

		return err
	}

	code := sub.Code

	task, err := s.taskRepo.GetById(ctx, sub.TaskID)
	if err != nil {
		if errors.Is(err, errs.ErrTaskNotFound) {
			return errs.ErrNotFound
		}

		return err
	}

	lang, err := s.langRepo.GetById(ctx, sub.LanguageID)
	if err != nil {
		if errors.Is(err, errs.ErrLanguageNotFound) {
			return errs.ErrNotFound
		}

		return err
	}

	langConfig, ok := config.Registry[lang.Slug]

	if !ok {
		return fmt.Errorf("Undefined Language")
	}

	testCasesId, err := s.subRepo.GetAllTestsIdForSubmission(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrSubmissionNotFound) {
			return errs.ErrNotFound
		}

		return err
	}

	if len(testCasesId) == 0 {
		slog.Warn("judge submission failed", slog.Any("err", errs.ErrEmptySlice))
		return errs.ErrEmptySlice
	}

	testCases, err := s.testRepo.GetAllTestsByIdSlice(ctx, testCasesId)
	if err != nil {
		if errors.Is(err, errs.ErrTestNotFound) {
			return errs.ErrNotFound
		}

		return err
	}

	compileExeRes, err := s.runner.Compile(ctx, langConfig, code)

	var maxTime, maxMem = 0, 0

	if err != nil || compileExeRes == nil || compileExeRes.ExitCode != 0 {
		return s.subRepo.SaveById(ctx, id, "CE", maxTime, maxMem)
	}

	for _, test := range testCases {
		runExeRes, err := s.runner.RunTest(
			ctx,
			langConfig,
			test.InputData,
			compileExeRes.Binary,
			float64(task.TimeLimit)/1000,
			int64(task.MemoryLimit),
		)

		if err != nil {
			slog.Error("Runtime system err", slog.Any("error", err))
			return s.subRepo.SaveById(ctx, id, "SE", maxTime, maxMem)
		}

		maxTime = max(maxTime, int(runExeRes.Duration))
		maxMem = max(maxMem, int(runExeRes.MemoryUsed))

		if runExeRes.TLE {
			return s.subRepo.SaveById(ctx, id, "TLE", maxTime, maxMem)

		} else if runExeRes.MLE {
			return s.subRepo.SaveById(ctx, id, "MLE", maxTime, maxMem)

		} else if runExeRes.ExitCode != 0 {
			return s.subRepo.SaveById(ctx, id, "RE", maxTime, maxMem)

		} else if strings.TrimSpace(test.OutputData) != strings.TrimSpace(runExeRes.Stdout) {
			return s.subRepo.SaveById(ctx, id, "WA", maxTime, maxMem)
		}

	}

	return s.subRepo.SaveById(ctx, id, "OK", maxTime, maxMem)

}

func NewJudgeService(
	subRepo JudgeSubmissionRepository,
	langRepo JudgeLanguageRepository,
	taskRepo JudgeTaskRepository,
	testRepo JudgeTestRepository,
	runner domain.CodeRunner) JudgeService {

	return &judgeService{
		subRepo:  subRepo,
		langRepo: langRepo,
		taskRepo: taskRepo,
		testRepo: testRepo,
		runner:   runner}
}
