package judge

import (
	"context"
	"judge-system/internal/judge/config"
	"judge-system/internal/judge/domain"
	"judge-system/internal/models"
)

//

type LangMock struct {
	GetByIdFunc func(ctx context.Context, id int) (*models.Language, error)
}

func (m *LangMock) GetById(ctx context.Context, id int) (*models.Language, error) {
	if m.GetByIdFunc != nil {
		return m.GetByIdFunc(ctx, id)
	}
	return nil, nil
}

type TaskMock struct {
	GetByIdFunc func(ctx context.Context, id int64) (*models.Task, error)
}

func (m *TaskMock) GetById(ctx context.Context, id int64) (*models.Task, error) {
	if m.GetByIdFunc != nil {
		return m.GetByIdFunc(ctx, id)
	}
	return nil, nil
}

//

type TestMock struct {
	GetAllTestsByIdSliceFunc func(ctx context.Context, testsId []int64) ([]models.TestCase, error)
}

func (m *TestMock) GetAllTestsByIdSlice(ctx context.Context, testsId []int64) ([]models.TestCase, error) {
	if m.GetAllTestsByIdSliceFunc != nil {
		return m.GetAllTestsByIdSliceFunc(ctx, testsId)
	}
	return nil, nil
}

//

type SubMock struct {
	GetAllTestsIdForSubmissionFunc func(ctx context.Context, id int64) ([]int64, error)
	GetByIdFunc                    func(ctx context.Context, id int64) (*models.Submission, error)
	SaveByIdFunc                   func(ctx context.Context, id int64, finalVerdict string, maxTime, maxMem int) error
}

func (m *SubMock) GetAllTestsIdForSubmission(ctx context.Context, id int64) ([]int64, error) {
	if m.GetAllTestsIdForSubmissionFunc != nil {
		return m.GetAllTestsIdForSubmissionFunc(ctx, id)
	}
	return nil, nil
}

func (m *SubMock) GetById(ctx context.Context, id int64) (*models.Submission, error) {
	if m.GetByIdFunc != nil {
		return m.GetByIdFunc(ctx, id)
	}
	return nil, nil
}

func (m *SubMock) SaveById(ctx context.Context, id int64, finalVerdict string, maxTime, maxMem int) error {
	if m.SaveByIdFunc != nil {
		return m.SaveByIdFunc(ctx, id, finalVerdict, maxTime, maxMem)
	}
	return nil
}

//

type RunnerMock struct {
	CompileFunc func(ctx context.Context, lang config.LanguageConfig, code string) (*domain.ExecutionResult, error)

	RunTestFunc func(
		ctx context.Context,
		lang config.LanguageConfig,
		input string,
		codeOrBin []byte,
		timeLimSec float64,
		memoryLimMb int64) (*domain.ExecutionResult, error)
}

func (m *RunnerMock) Compile(ctx context.Context, lang config.LanguageConfig, code string) (*domain.ExecutionResult, error) {
	if m.CompileFunc != nil {
		return m.CompileFunc(ctx, lang, code)
	}

	return nil, nil
}

func (m *RunnerMock) RunTest(
	ctx context.Context,
	lang config.LanguageConfig,
	input string,
	codeOrBin []byte,
	timeLimSec float64,
	memoryLimMb int64) (*domain.ExecutionResult, error) {

	if m.RunTestFunc != nil {
		return m.RunTestFunc(ctx, lang, input, codeOrBin, timeLimSec, memoryLimMb)
	}

	return nil, nil
}

//

func setupBaseSuccessMocks(s *SubMock, l *LangMock, t *TaskMock, te *TestMock) {
	s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
		return &models.Submission{ID: id, UserID: 67, TaskID: 5, LanguageID: 1, Code: "fmt.Println()"}, nil
	}
	t.GetByIdFunc = func(ctx context.Context, id int64) (*models.Task, error) {
		return &models.Task{ID: 5, TimeLimit: 1000, MemoryLimit: 256}, nil
	}
	l.GetByIdFunc = func(ctx context.Context, id int) (*models.Language, error) {
		return &models.Language{Slug: "golang"}, nil
	}
	s.GetAllTestsIdForSubmissionFunc = func(ctx context.Context, id int64) ([]int64, error) {
		return []int64{10, 11}, nil
	}
	te.GetAllTestsByIdSliceFunc = func(ctx context.Context, testsId []int64) ([]models.TestCase, error) {
		return []models.TestCase{
			{ID: 10, InputData: "1", OutputData: "2"},
			{ID: 11, InputData: "3", OutputData: "4"},
		}, nil
	}
}
