package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type cppRunner struct{}

func NewCppRunner() LanguageRunner {
	return &cppRunner{}
}

func (r *cppRunner) Compile(ctx context.Context, source_code string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "cpp_build_*")

	if err != nil {
		return "", err
	}

	binPath := filepath.Join(tmpDir, "solution")
	sourcePath := filepath.Join(tmpDir, "solution.cpp")

	if err := os.WriteFile(sourcePath, []byte(source_code), 0644); err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "g++", "-O2", sourcePath, "-o", binPath)

	if out, err := cmd.Output(); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("Error with compilation %s", string(out))
	}

	return binPath, nil
}

func (r *cppRunner) Run(ctx context.Context, execPath, inputData string, time_limit, memory_limit, boxID int) (*TaskResult, error) {
	return RunIsolation(ctx, execPath, inputData, time_limit, memory_limit, boxID, []string{"./solution"})
}

func (r *cppRunner) CleanUp(execPath string) {
	os.RemoveAll(filepath.Dir(execPath))
}
