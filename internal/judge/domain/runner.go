package domain

import (
	"context"
	"judge-system/internal/judge/config"
)

type CodeRunner interface {
	Compile(ctx context.Context, lang config.LanguageConfig, code string) (*ExecutionResult, error)

	RunTest(
		ctx context.Context,
		lang config.LanguageConfig,
		input string,
		codeOrBin []byte,
		timeLimSec float64,
		memoryLimMb int64) (*ExecutionResult, error)
}
