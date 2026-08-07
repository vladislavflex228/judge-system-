package service

import (
	"context"
	"errors"
	"fmt"
	"judge-system/internal/errs"
	"judge-system/internal/models"
	"judge-system/internal/runner"
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
	subRepo       JudgeSubmissionRepository
	langRepo      JudgeLanguageRepository
	taskRepo      JudgeTaskRepository
	testRepo      JudgeTestRepository
	runnerManager *runner.RunnerManager
}

func (s *judgeService) JudgeSubmission(ctx context.Context, id int64) error {

	if id < 0 {
		return errs.ErrInvalidInput
	}

	testCasesId, err := s.subRepo.GetAllTestsIdForSubmission(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrSubmissionNotFound) {
			return errs.ErrNotFound
		}

		return err
	}

	if len(testCasesId) == 0 {
		slog.Warn("judge submission failes", slog.Any("err", errs.ErrEmptySlice))
		return errs.ErrEmptySlice
	}

	slog.Info("Дошли до get all tests by id slice", "slice", testCasesId)

	testCases, err := s.testRepo.GetAllTestsByIdSlice(ctx, testCasesId)
	if err != nil {
		if errors.Is(err, errs.ErrTestNotFound) {
			return errs.ErrNotFound
		}

		return err
	}

	slog.Info("Дошли до sub get", "slice", testCases)

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

	language_id := sub.LanguageID
	language, err := s.langRepo.GetById(ctx, language_id)
	if err != nil {
		if errors.Is(err, errs.ErrLanguageNotFound) {
			return errs.ErrNotFound
		}

		return err
	}

	langRunner, ok := s.runnerManager.Runners[language.Slug]

	if !ok {
		return fmt.Errorf("unsupported language %s", language.Slug)
	}

	path, err := langRunner.Compile(ctx, code)

	if err != nil {
		return fmt.Errorf("Submission id : %d , compilation error : %w", id, err)
	}

	defer langRunner.CleanUp(path)

	boxID := 0
	var maxTime int
	var maxMem int
	finalVerdict := "OK"

	for _, test := range testCases {
		res, err := langRunner.Run(ctx, path, test.InputData, task.TimeLimit, task.MemoryLimit, boxID)
		boxID++

		if err != nil {
			finalVerdict = "SE"
			break
		}

		maxTime = max(res.TimeUsed, task.TimeLimit)
		maxMem = max(res.MemoryUsed, task.MemoryLimit)

		if res.Verdict != "OK" {
			finalVerdict = res.Verdict
			break
		}

		if strings.TrimSpace(res.Stdout) != strings.TrimSpace(test.OutputData) {
			finalVerdict = "WA"
			break
		}

	}

	return s.subRepo.SaveById(ctx, id, finalVerdict, maxTime, maxMem)

}

func NewJudgeService(subRepo JudgeSubmissionRepository, langRepo JudgeLanguageRepository, taskRepo JudgeTaskRepository, testRepo JudgeTestRepository, runnerManager *runner.RunnerManager) JudgeService {
	return &judgeService{
		subRepo:       subRepo,
		langRepo:      langRepo,
		taskRepo:      taskRepo,
		testRepo:      testRepo,
		runnerManager: runnerManager}
}
