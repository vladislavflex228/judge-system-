package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

	defer r.cli.ContainerRemove(context.Background(), resp.ID, client.ContainerRemoveOptions{Force: true})

	tarReader, err := createTarArchive(lang.SourceFile, []byte(code))

	_, err = r.cli.CopyToContainer(ctx, resp.ID, client.CopyToContainerOptions{
		DestinationPath: r.confBuilder.WorkDir,
		Content:         tarReader,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to copy content to container :%w", err)
	}

	startTime := time.Now()
	if _, err := r.cli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	var exitCode int
	waitRes := r.cli.ContainerWait(ctx, resp.ID, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})
	duration := time.Since(startTime)

	select {
	case st := <-waitRes.Result:
		exitCode = int(st.StatusCode)
	case err := <-waitRes.Error:
		return nil, fmt.Errorf("compile wait error: %w", err)
	}

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
		binPath := r.confBuilder.WorkDir + "/" + lang.BinaryFile
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

	contConf, hostConf := r.confBuilder.BuildRunCmd(lang.DockerImage, lang.RunCmd, timeLimSec, memoryLimMb)

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

	tarReader, err := createTarArchive(targetFileName, codeOrBin)

	if err != nil {
		return nil, fmt.Errorf("failed to archive content: %w", err)
	}

	_, err = r.cli.CopyToContainer(ctx, resp.ID, client.CopyToContainerOptions{
		DestinationPath: r.confBuilder.WorkDir,
		Content:         tarReader})

	if err != nil {
		return nil, fmt.Errorf("failed to copy content to run container, %w", err)
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

	startTime := time.Now()
	_, err = r.cli.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to start run container: %w", err)
	}

	go func() {
		type closeWriter interface {
			CloseWrite() error
		}

		if cw, ok := attachResp.Conn.(closeWriter); ok {
			_ = cw.CloseWrite()
		} else {
			attachResp.Close()
		}
	}()

	var outBuf, errBuf bytes.Buffer
	go func() {
		_, _ = stdcopy.StdCopy(&outBuf, &errBuf, attachResp.Reader)
	}()

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
