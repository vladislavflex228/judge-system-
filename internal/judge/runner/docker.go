package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"judge-system/internal/judge/config"
	"judge-system/internal/judge/domain"
	"log/slog"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type DockerRunner struct {
	cli         *client.Client
	confBuilder *config.ConfigBuilder
}

func NewDockerRunner(confBuilder *config.ConfigBuilder) (*DockerRunner, error) {
	cli, err := client.New(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &DockerRunner{cli: cli, confBuilder: confBuilder}, nil
}

func (r *DockerRunner) Compile(ctx context.Context, lang config.LanguageConfig, code string) (*domain.ExecutionResult, error) {
	contConfig, hostConfig := r.confBuilder.BuildCompileCmd(lang.DockerImage, lang.CompileCmd)

	resp, err := r.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     contConfig,
		HostConfig: hostConfig,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create compile container :%w", err)
	}

	defer r.cli.ContainerRemove(context.Background(), resp.ID, client.ContainerRemoveOptions{Force: true}) //принудительное удаление контейнера(без остановки).Без флага может быть ошибка,если контейнер еще работает
	//
	tarReader, err := createTarArchive(lang.SourceFile, []byte(code))

	attach, err := r.cli.ContainerAttach(ctx, resp.ID, client.ContainerAttachOptions{Stream: true, Stdin: true})
	if err != nil {
		return nil, fmt.Errorf("ошибка attach: %w", err)
	}

	go func() {
		defer attach.Close()
		_, _ = io.Copy(attach.Conn, tarReader)
		attach.CloseWrite()
	}()
	//
	startTime := time.Now()
	if _, err := r.cli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container :%w", err)
	}

	var exitCode int
	execCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	waitRes := r.cli.ContainerWait(execCtx, resp.ID, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})
	select {
	case st := <-waitRes.Result:
		exitCode = int(st.StatusCode)
	case err := <-waitRes.Error:
		return nil, fmt.Errorf("compile wait error: %w", err)
	}

	slog.Info("check", slog.Int("exitcode", exitCode))

	duration := time.Since(startTime)

	logsResp, err := r.cli.ContainerLogs(ctx, resp.ID, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true})

	if err != nil {
		return nil, err
	}

	defer logsResp.Close()

	var outBuf, errBuf bytes.Buffer

	_, _ = stdcopy.StdCopy(&outBuf, &errBuf, logsResp)

	var binaryBytes []byte

	if exitCode == 0 && lang.BinaryFile != "" {
		binPath := "/tmp" + "/" + lang.BinaryFile
		binaryBytes, err = r.ExtractBinary(ctx, resp.ID, binPath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract compiled binary: %w ", err)
		}
	}

	return &domain.ExecutionResult{
		ExitCode: exitCode,
		Stdout:   outBuf.String(),
		Stderr:   errBuf.String(),
		Duration: duration,
		Binary:   binaryBytes,
	}, nil
}

func (r *DockerRunner) RunTest(
	ctx context.Context,
	lang config.LanguageConfig,
	input string,
	codeOrBin []byte,
	timeLimSec float64,
	memoryLimMb int64) (*domain.ExecutionResult, error) {

	contConf, hostConf := r.confBuilder.BuildRunCmd(lang.DockerImage, lang.RunCmd, 1, memoryLimMb)

	resp, err := r.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     contConf,
		HostConfig: hostConf,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create run container : %w", err)
	}

	defer func() {
		_, _ = r.cli.ContainerRemove(context.Background(), resp.ID, client.ContainerRemoveOptions{Force: true})
	}()

	targetFileName := lang.BinaryFile

	if targetFileName == "" {
		targetFileName = lang.SourceFile //для интерпретируемых языков
	}

	tarReader, err := createTarArchiveForRun(targetFileName, []byte(input), codeOrBin)

	if err != nil {
		return nil, fmt.Errorf("failed to archive content: %w", err)
	}

	attachResp, err := r.cli.ContainerAttach(ctx, resp.ID, client.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true})

	if err != nil {
		return nil, fmt.Errorf("failed to attach to container,%w : ", err)
	}

	defer attachResp.Close()

	go func() {
		_, err := io.Copy(attachResp.Conn, tarReader)
		if err != nil {
			slog.Error("failed to copy tar to container stdin", slog.Any("error", err))
		}
		_ = attachResp.CloseWrite()
	}()

	var outBuf, errBuf bytes.Buffer
	go func() {
		_, _ = stdcopy.StdCopy(&outBuf, &errBuf, attachResp.Reader)
	}()

	startTime := time.Now()
	_, err = r.cli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start run container: %w", err)
	}

	tleDuration := time.Duration((timeLimSec + 0.05) * float64(time.Second))

	timeoutCtx, cancel := context.WithTimeout(ctx, tleDuration)

	defer cancel()

	waitResp := r.cli.ContainerWait(timeoutCtx, resp.ID, client.ContainerWaitOptions{
		Condition: container.WaitConditionNotRunning})

	result := &domain.ExecutionResult{}

	select {
	case err := <-waitResp.Error:
		return nil, fmt.Errorf("run wait error, %w", err)
	case <-timeoutCtx.Done():
		_, err = r.cli.ContainerKill(context.Background(), resp.ID, client.ContainerKillOptions{Signal: "SIGKILL"})

		if ctx.Err() != nil {
			return nil, fmt.Errorf("execution was canceled by client/server context: %w", err)
		}
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				result.TLE = true
			} else {
				return nil, fmt.Errorf("failed at killing run container %s , %w", resp.ID, err)
			}
		}
	case exitCode := <-waitResp.Result:
		result.ExitCode = int(exitCode.StatusCode)

		memUsageBytes, err := r.GetMemoryUsage(ctx, resp.ID)

		if err != nil {
			slog.Error("get memory usage", "error", err)
		}

		result.MemoryUsed = int64(memUsageBytes)
	}

	result.Duration = time.Since(startTime)

	inspectResp, err := r.cli.ContainerInspect(ctx, resp.ID, client.ContainerInspectOptions{})

	if err == nil {
		if inspectResp.Container.State.OOMKilled || inspectResp.Container.State.ExitCode == 137 {
			result.MLE = true
		}
	}

	result.Stdout = outBuf.String()
	result.Stderr = errBuf.String()

	return result, nil
}
