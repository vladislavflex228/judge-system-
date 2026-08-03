package service

import (
	"context"
	"fmt"
	"judge-system/internal/models"
	"judge-system/internal/runner"
	"strings"
)

type JudgeSubmissionRepository interface {
	GetAllTestsForSubmission(ctx context.Context, id int64) ([]models.TestCase, error)
	GetByID(ctx context.Context, id int64) (*models.Submission, error)
	SaveById(ctx context.Context, id int64, finalVerdict string, maxTime, maxMem int) error
}

type JudgeLanguageRepository interface {
	GetByID(ctx context.Context, id int) (*models.Language, error)
}

type JudgeTaskRepository interface {
	GetByID(ctx context.Context, id int64) (*models.Task, error)
}

type JudgeService interface {
	JudgeSubmission(ctx context.Context, id int64) error
}

type judgeService struct {
	subRepo       JudgeSubmissionRepository
	langRepo      JudgeLanguageRepository
	taskRepo      JudgeTaskRepository
	runnerManager *runner.RunnerManager
}

func (s *judgeService) JudgeSubmission(ctx context.Context, id int64) error {

	if id < 0 {
		return ErrInvalidInput
	}

	testCases, err := s.subRepo.GetAllTestsForSubmission(ctx, id)
	if err != nil {
		return err
	}

	sub, err := s.subRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	code := sub.Code

	task, err := s.taskRepo.GetByID(ctx, sub.TaskID)
	if err != nil {
		return err
	}

	language_id := sub.LanguageID
	language, err := s.langRepo.GetByID(ctx, language_id)
	if err != nil {
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

func NewJudgeService(subRepo JudgeSubmissionRepository, langRepo JudgeLanguageRepository, taskRepo JudgeTaskRepository, runnerManager *runner.RunnerManager) JudgeService {
	return &judgeService{
		subRepo:       subRepo,
		langRepo:      langRepo,
		taskRepo:      taskRepo,
		runnerManager: runnerManager}
}
