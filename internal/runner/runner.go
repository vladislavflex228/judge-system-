package runner

import (
	"context"
)

// Pattern Strategy

type TaskResult struct {
	Verdict    string
	Stdout     string
	Stderr     string
	TimeUsed   int
	MemoryUsed int
}

type LanguageRunner interface {
	Compile(ctx context.Context, source_code string) (string, error)

	Run(ctx context.Context, binPath, inputData string, time_limit, memory_limit, boxID int) (*TaskResult, error)

	CleanUp(execPath string)
}

type RunnerManager struct {
	Runners map[string]LanguageRunner
}

func NewRunnerManager(runners map[string]LanguageRunner) *RunnerManager {
	return &RunnerManager{
		Runners: runners}
}
