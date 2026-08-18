package judge

import (
	"context"
	"errors"
	"judge-system/internal/errs"
	"judge-system/internal/judge/config"
	"judge-system/internal/judge/domain"
	"judge-system/internal/middleware"
	"judge-system/internal/models"
	"testing"
)

func TestJudgeSubmissionPipelineErrors(t *testing.T) {

	validCtx := context.WithValue(context.Background(), middleware.UserIDKey, int64(67))

	testCases := []struct {
		testName        string
		subId           int64
		ctx             context.Context
		setupMocks      func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock)
		wantErr         error
		expectedVerdict string
	}{
		{
			testName: "1.Negative subId",
			subId:    -2,
			wantErr:  errs.ErrInvalidInput,
		},
		{
			testName: "2.submission GetById Not Found",
			subId:    1,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return nil, errs.ErrSubmissionNotFound
				}
			},
			wantErr: errs.ErrNotFound,
		},
		{
			testName: "3.submission GetById Bd error",
			subId:    1,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return nil, errs.ErrDataBase
				}
			},
			wantErr: errs.ErrDataBase,
		},
		{
			testName: "4.empty context value",
			subId:    1,
			ctx:      context.Background(),
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return &models.Submission{UserID: 100}, nil
				}
			},
			wantErr: errs.ErrWrongUserIDFormat,
		},
		{
			testName: "5.user_id discrepancy",
			subId:    1,
			ctx:      validCtx,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return &models.Submission{UserID: 100}, nil
				}
			},
			wantErr: errs.ErrUserDiscrepancy,
		},
		{
			testName: "6.task GetById Not Found",
			subId:    1,
			ctx:      validCtx,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return &models.Submission{UserID: 67}, nil
				}
				t.GetByIdFunc = func(ctx context.Context, id int64) (*models.Task, error) {
					return nil, errs.ErrTaskNotFound
				}
			},
			wantErr: errs.ErrNotFound,
		},
		{
			testName: "7.task GetById Bd error",
			subId:    1,
			ctx:      validCtx,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return &models.Submission{UserID: 67}, nil
				}
				t.GetByIdFunc = func(ctx context.Context, id int64) (*models.Task, error) {
					return nil, errs.ErrDataBase
				}
			},
			wantErr: errs.ErrDataBase,
		},
		{
			testName: "8.language GetById Not Found",
			subId:    1,
			ctx:      validCtx,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return &models.Submission{UserID: 67}, nil
				}
				t.GetByIdFunc = func(ctx context.Context, id int64) (*models.Task, error) {
					return &models.Task{ID: 1}, nil
				}
				l.GetByIdFunc = func(ctx context.Context, id int) (*models.Language, error) {
					return nil, errs.ErrLanguageNotFound
				}
			},
			wantErr: errs.ErrNotFound,
		},
		{
			testName: "9.language GetById Bd error",
			subId:    1,
			ctx:      validCtx,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return &models.Submission{UserID: 67}, nil
				}
				t.GetByIdFunc = func(ctx context.Context, id int64) (*models.Task, error) {
					return &models.Task{ID: 1}, nil
				}
				l.GetByIdFunc = func(ctx context.Context, id int) (*models.Language, error) {
					return nil, errs.ErrDataBase
				}
			},
			wantErr: errs.ErrDataBase,
		},
		{
			testName: "10.Undefined language",
			subId:    1,
			ctx:      validCtx,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return &models.Submission{UserID: 67}, nil
				}
				t.GetByIdFunc = func(ctx context.Context, id int64) (*models.Task, error) {
					return &models.Task{ID: 1}, nil
				}
				l.GetByIdFunc = func(ctx context.Context, id int) (*models.Language, error) {
					return &models.Language{Slug: "ruby"}, nil
				}
			},
			wantErr: errs.ErrUndefinedLanguage,
		},
		{
			testName: "11.submission GetAllTestsForSubmission not found",
			subId:    1,
			ctx:      validCtx,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return &models.Submission{UserID: 67}, nil
				}
				s.GetAllTestsIdForSubmissionFunc = func(ctx context.Context, id int64) ([]int64, error) {
					return nil, errs.ErrSubmissionNotFound
				}
				t.GetByIdFunc = func(ctx context.Context, id int64) (*models.Task, error) {
					return &models.Task{ID: 1}, nil
				}
				l.GetByIdFunc = func(ctx context.Context, id int) (*models.Language, error) {
					return &models.Language{Slug: "golang"}, nil
				}
			},
			wantErr: errs.ErrNotFound,
		},
		{
			testName: "12.submission GetAllTestsForSubmission bd error",
			subId:    1,
			ctx:      validCtx,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return &models.Submission{UserID: 67}, nil
				}
				s.GetAllTestsIdForSubmissionFunc = func(ctx context.Context, id int64) ([]int64, error) {
					return nil, errs.ErrDataBase
				}
				t.GetByIdFunc = func(ctx context.Context, id int64) (*models.Task, error) {
					return &models.Task{ID: 1}, nil
				}
				l.GetByIdFunc = func(ctx context.Context, id int) (*models.Language, error) {
					return &models.Language{Slug: "golang"}, nil
				}
			},
			wantErr: errs.ErrDataBase,
		},
		{
			testName: "13.empty test slice",
			subId:    1,
			ctx:      validCtx,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return &models.Submission{UserID: 67}, nil
				}
				s.GetAllTestsIdForSubmissionFunc = func(ctx context.Context, id int64) ([]int64, error) {
					return []int64{}, nil
				}
				t.GetByIdFunc = func(ctx context.Context, id int64) (*models.Task, error) {
					return &models.Task{ID: 1}, nil
				}
				l.GetByIdFunc = func(ctx context.Context, id int) (*models.Language, error) {
					return &models.Language{Slug: "golang"}, nil
				}
			},
			wantErr: errs.ErrEmptySlice,
		},
		{
			testName: "14.test GetById Not Found",
			subId:    1,
			ctx:      validCtx,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return &models.Submission{UserID: 67}, nil
				}
				s.GetAllTestsIdForSubmissionFunc = func(ctx context.Context, id int64) ([]int64, error) {
					return []int64{1, 2, 3}, nil
				}
				t.GetByIdFunc = func(ctx context.Context, id int64) (*models.Task, error) {
					return &models.Task{ID: 1}, nil
				}
				l.GetByIdFunc = func(ctx context.Context, id int) (*models.Language, error) {
					return &models.Language{Slug: "golang"}, nil
				}
				te.GetAllTestsByIdSliceFunc = func(ctx context.Context, testsId []int64) ([]models.TestCase, error) {
					return nil, errs.ErrTestNotFound
				}
			},
			wantErr: errs.ErrNotFound,
		},
		{
			testName: "15.test GetById Bd error",
			subId:    1,
			ctx:      validCtx,
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock) {
				s.GetByIdFunc = func(ctx context.Context, id int64) (*models.Submission, error) {
					return &models.Submission{UserID: 67}, nil
				}
				s.GetAllTestsIdForSubmissionFunc = func(ctx context.Context, id int64) ([]int64, error) {
					return []int64{1, 2, 3}, nil
				}
				t.GetByIdFunc = func(ctx context.Context, id int64) (*models.Task, error) {
					return &models.Task{ID: 1}, nil
				}
				l.GetByIdFunc = func(ctx context.Context, id int) (*models.Language, error) {
					return &models.Language{Slug: "golang"}, nil
				}
				te.GetAllTestsByIdSliceFunc = func(ctx context.Context, testsId []int64) ([]models.TestCase, error) {
					return nil, errs.ErrDataBase
				}
			},
			wantErr: errs.ErrDataBase,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			mockSubRepo := &SubMock{}
			mockLangRepo := &LangMock{}
			mockTaskRepo := &TaskMock{}
			mockTestRepo := &TestMock{}
			mockRunner := &RunnerMock{}

			if tt.setupMocks != nil {
				tt.setupMocks(mockSubRepo, mockTestRepo, mockTaskRepo, mockLangRepo)
			}

			service := NewJudgeService(mockSubRepo, mockLangRepo, mockTaskRepo, mockTestRepo, mockRunner, config.Registry)

			_, gotErr := service.JudgeSubmission(tt.ctx, tt.subId)

			if tt.wantErr != gotErr {
				t.Fatalf("wrong error, expected : %v,got : %v", tt.wantErr, gotErr)
			}
		})
	}
}

func TestJudgeSubmissionVerdictAndSaving(t *testing.T) {
	validCtx := context.WithValue(context.Background(), middleware.UserIDKey, int64(67))
	testCases := []struct {
		testName      string
		setupMocks    func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock, r *RunnerMock)
		expectVerdict string
		expectMaxTime int
		expectMaxMem  int
	}{
		{
			testName: "CE Verdict",
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock, r *RunnerMock) {
				setupBaseSuccessMocks(s, l, t, te)
				r.CompileFunc = func(ctx context.Context, lang config.LanguageConfig, code string) (*domain.ExecutionResult, error) {
					return &domain.ExecutionResult{ExitCode: 1, Stderr: "syntax err"}, nil
				}
			},
			expectVerdict: "CE",
			expectMaxTime: 0,
			expectMaxMem:  0,
		},
		{
			testName: "TLE Verdict",
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock, r *RunnerMock) {
				setupBaseSuccessMocks(s, l, t, te)
				r.CompileFunc = func(ctx context.Context, lang config.LanguageConfig, code string) (*domain.ExecutionResult, error) {
					return &domain.ExecutionResult{ExitCode: 0, Binary: []byte("compilation result")}, nil
				}

				testCount := 0

				r.RunTestFunc = func(ctx context.Context, lang config.LanguageConfig, input string, codeOrBin []byte, timeLimSec float64, memoryLimMb int64) (*domain.ExecutionResult, error) {
					testCount++
					if testCount == 1 {
						return &domain.ExecutionResult{ExitCode: 0, Duration: 15, MemoryUsed: 50, Stdout: "2"}, nil
					}
					return &domain.ExecutionResult{TLE: true, Duration: 1000, MemoryUsed: 40}, nil

				}
			},
			expectVerdict: "TLE",
			expectMaxTime: 1000,
			expectMaxMem:  50,
		},
		{
			testName: "SE Verdict",
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock, r *RunnerMock) {
				setupBaseSuccessMocks(s, l, t, te)
				r.CompileFunc = func(ctx context.Context, lang config.LanguageConfig, code string) (*domain.ExecutionResult, error) {
					return &domain.ExecutionResult{ExitCode: 0, Binary: []byte("compilation result")}, nil
				}

				r.RunTestFunc = func(ctx context.Context, lang config.LanguageConfig, input string, codeOrBin []byte, timeLimSec float64, memoryLimMb int64) (*domain.ExecutionResult, error) {
					return nil, errors.New("System error")
				}
			},
			expectVerdict: "SE",
			expectMaxTime: 0,
			expectMaxMem:  0,
		},
		{
			testName: "MLE Verdict",
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock, r *RunnerMock) {
				setupBaseSuccessMocks(s, l, t, te)
				r.CompileFunc = func(ctx context.Context, lang config.LanguageConfig, code string) (*domain.ExecutionResult, error) {
					return &domain.ExecutionResult{ExitCode: 0, Binary: []byte("compilation result")}, nil
				}
				testCount := 0
				r.RunTestFunc = func(ctx context.Context, lang config.LanguageConfig, input string, codeOrBin []byte, timeLimSec float64, memoryLimMb int64) (*domain.ExecutionResult, error) {
					testCount++
					if testCount == 1 {
						return &domain.ExecutionResult{ExitCode: 0, Stdout: "2", Duration: 100, MemoryUsed: 200}, nil
					}

					return &domain.ExecutionResult{MLE: true, ExitCode: 0, Duration: 200, MemoryUsed: 256}, nil
				}
			},
			expectVerdict: "MLE",
			expectMaxTime: 200,
			expectMaxMem:  256,
		},
		{
			testName: "RE Verdict",
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock, r *RunnerMock) {
				setupBaseSuccessMocks(s, l, t, te)
				r.CompileFunc = func(ctx context.Context, lang config.LanguageConfig, code string) (*domain.ExecutionResult, error) {
					return &domain.ExecutionResult{ExitCode: 0, Binary: []byte("compilation result")}, nil
				}
				r.RunTestFunc = func(ctx context.Context, lang config.LanguageConfig, input string, codeOrBin []byte, timeLimSec float64, memoryLimMb int64) (*domain.ExecutionResult, error) {
					return &domain.ExecutionResult{ExitCode: 2, Duration: 200, MemoryUsed: 100}, nil
				}
			},
			expectVerdict: "RE",
			expectMaxTime: 200,
			expectMaxMem:  100,
		},
		{
			testName: "WA Verdict",
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock, r *RunnerMock) {
				setupBaseSuccessMocks(s, l, t, te)
				r.CompileFunc = func(ctx context.Context, lang config.LanguageConfig, code string) (*domain.ExecutionResult, error) {
					return &domain.ExecutionResult{ExitCode: 0, Binary: []byte("compilation result")}, nil
				}
				r.RunTestFunc = func(ctx context.Context, lang config.LanguageConfig, input string, codeOrBin []byte, timeLimSec float64, memoryLimMb int64) (*domain.ExecutionResult, error) {
					return &domain.ExecutionResult{ExitCode: 0, Stdout: "wrong ans", Duration: 80, MemoryUsed: 15}, nil
				}
			},
			expectVerdict: "WA",
			expectMaxTime: 80,
			expectMaxMem:  15,
		},
		{
			testName: "OK Verdict",
			setupMocks: func(s *SubMock, te *TestMock, t *TaskMock, l *LangMock, r *RunnerMock) {
				setupBaseSuccessMocks(s, l, t, te)
				r.CompileFunc = func(ctx context.Context, lang config.LanguageConfig, code string) (*domain.ExecutionResult, error) {
					return &domain.ExecutionResult{ExitCode: 0, Binary: []byte("compilation result")}, nil
				}
				testCount := 0
				r.RunTestFunc = func(ctx context.Context, lang config.LanguageConfig, input string, codeOrBin []byte, timeLimSec float64, memoryLimMb int64) (*domain.ExecutionResult, error) {
					testCount++
					if testCount == 1 {
						return &domain.ExecutionResult{ExitCode: 0, Stdout: "2", Duration: 80, MemoryUsed: 15}, nil
					}

					return &domain.ExecutionResult{ExitCode: 0, Stdout: "4", Duration: 100, MemoryUsed: 10}, nil
				}
			},
			expectVerdict: "OK",
			expectMaxTime: 100,
			expectMaxMem:  15,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.testName, func(t *testing.T) {
			var (
				capturedId      int64
				capturedVerdict string
				capturedMaxTime int
				capturedMaxMem  int
				saveCalled      bool
			)

			mockSubRepo := &SubMock{
				SaveByIdFunc: func(ctx context.Context, id int64, finalVerdict string, maxTime, maxMem int) error {
					saveCalled = true
					capturedId = id
					capturedVerdict = finalVerdict
					capturedMaxTime = maxTime
					capturedMaxMem = maxMem
					return nil
				},
			}

			mockLangRepo := &LangMock{}
			mockTaskRepo := &TaskMock{}
			mockTestRepo := &TestMock{}
			mockRunner := &RunnerMock{}
			langRegistry := make(map[string]config.LanguageConfig)
			langRegistry["golang"] = config.LanguageConfig{}

			tt.setupMocks(mockSubRepo, mockTestRepo, mockTaskRepo, mockLangRepo, mockRunner)

			service := NewJudgeService(mockSubRepo, mockLangRepo, mockTaskRepo, mockTestRepo, mockRunner, langRegistry)

			_, _ = service.JudgeSubmission(validCtx, 1)

			if !saveCalled {
				t.Fatal("informaion has't saved")
			}

			if capturedId != 1 {
				t.Fatalf("submission id was rubbed out, expected: %d, got: %d", 1, capturedId)
			}

			if capturedVerdict != tt.expectVerdict {
				t.Fatalf("wrong verdict, expected : %s,got :%s", tt.expectVerdict, capturedVerdict)
			}

			if capturedMaxTime != tt.expectMaxTime {
				t.Fatalf("wrong max time, expected : %d,got :%d", tt.expectMaxTime, capturedMaxTime)
			}

			if capturedMaxMem != tt.expectMaxMem {
				t.Fatalf("wrong max mem, expected : %d,got :%d", tt.expectMaxMem, capturedMaxMem)
			}

		})
	}
}
