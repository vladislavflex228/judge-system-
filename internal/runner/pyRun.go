package runner

import (
	"context"
	"os"
	"path/filepath"
)

type pyRunner struct{}

func NewPyRunner() LanguageRunner {
	return &cppRunner{}
}

func (r *pyRunner) Compile(ctx context.Context, source_code string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "py_src_*")

	if err != nil {
		return "", nil
	}

	sourcePath := filepath.Join(tmpDir, "solution.py")

	if err := os.WriteFile(sourcePath, []byte(source_code), 0644); err != nil {
		return "", nil
	}

	return sourcePath, nil
}

func (r *pyRunner) Run(ctx context.Context, execPath, inputData string, time_limit, memory_limit, boxID int) (*TaskResult, error) {
	return RunIsolation(ctx, execPath, inputData, time_limit, memory_limit, boxID, []string{"/usr/bin/python3", "solution.py"})
}
