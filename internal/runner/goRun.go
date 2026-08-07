package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type goRunner struct{}

func NewGoRunner() LanguageRunner {
	return &goRunner{}
}

func (r *goRunner) Compile(ctx context.Context, source_code string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "go_build_*")

	if err != nil {
		return "", err
	}

	sourcePath := filepath.Join(tmpDir, "solution.go")
	binPath := filepath.Join(tmpDir, "solution")

	if err := os.WriteFile(sourcePath, []byte(source_code), 0644); err != nil {
		return "", err
	}

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, sourcePath)

	if _, err := cmd.Output(); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("Error with compilation %s", err.Error())
	}

	return binPath, nil
}

func (r *goRunner) Run(ctx context.Context, execPath, inputData string, time_limit, memory_limit, boxID int) (*TaskResult, error) {
	return RunIsolation(ctx, execPath, inputData, time_limit, memory_limit, boxID, []string{"./solution"})
}

func (r *goRunner) CleanUp(execPath string) {
	os.RemoveAll(filepath.Dir(execPath))
}
