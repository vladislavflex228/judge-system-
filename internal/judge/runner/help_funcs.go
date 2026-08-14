package runner

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/client"
)

type MemoryStats struct {
	Usage    uint64 `json:"usage"`
	MaxUsage uint64 `json:"max_usage"`
	Stats    struct {
		Peak     uint64 `json:"peak"`
		Anon     uint64 `json:"anon"`
		Inactive uint64 `json:"inactive_file"`
	} `json:"stats"`
}

type ContainerStatsResponse struct {
	MemoryStats MemoryStats `json:"memory_stats"`
}

func (r *DockerRunner) GetMemoryUsage(ctx context.Context, containerID string) (uint64, error) {
	stats, err := r.cli.ContainerStats(ctx, containerID, client.ContainerStatsOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get stats: %w", err)
	}
	defer stats.Body.Close()

	var statsJSON ContainerStatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		return 0, fmt.Errorf("failed to decode stats json: %w", err)
	}

	maxMemory := statsJSON.MemoryStats.MaxUsage
	if maxMemory == 0 {
		maxMemory = statsJSON.MemoryStats.Stats.Peak
	}

	cleanMemory := maxMemory - statsJSON.MemoryStats.Stats.Inactive

	return cleanMemory, nil
}

func createTarArchive(filename string, content []byte) (io.Reader, error) {
	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)

	header := &tar.Header{
		Name:    filename,
		Size:    int64(len(content)),
		Mode:    0644,
		ModTime: time.Now(),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return nil, err
	}

	if _, err := tarWriter.Write(content); err != nil {
		return nil, err
	}

	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	return &buf, nil
}

func createTarArchiveForRun(bin_file_name string, input_content, bin_content []byte) (io.Reader, error) {
	var buf bytes.Buffer

	tarWriter := tar.NewWriter(&buf)

	header := &tar.Header{
		Name:    bin_file_name,
		Size:    int64(len(bin_content)),
		Mode:    0755,
		ModTime: time.Now(),
	}

	err := tarWriter.WriteHeader(header)

	if err != nil {
		return nil, err
	}

	_, err = tarWriter.Write(bin_content)

	if err != nil {
		return nil, err
	}

	input_header := &tar.Header{
		Name:    "input.txt",
		Size:    int64(len(input_content)),
		Mode:    0644,
		ModTime: time.Now(),
	}

	err = tarWriter.WriteHeader(input_header)

	if err != nil {
		return nil, err
	}

	_, err = tarWriter.Write(input_content)

	if err != nil {
		return nil, err
	}

	err = tarWriter.Close()

	if err != nil {
		return nil, err
	}

	return &buf, nil

}

func (r *DockerRunner) ExtractBinary(ctx context.Context, container_id string, binaryPath string) ([]byte, error) {
	copyRes, err := r.cli.CopyFromContainer(ctx, container_id, client.CopyFromContainerOptions{SourcePath: binaryPath})

	if err != nil {
		return nil, fmt.Errorf("failed at copy from compile container : %w", err)
	}

	defer copyRes.Content.Close()

	tarReader := tar.NewReader(copyRes.Content)

	for {
		header, err := tarReader.Next() //tarReader передвигает курсор для чтения на 512 байт(перепрыгивает чере заголовок)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed at read tar header : %w", err)
		}

		if header.Typeflag == tar.TypeReg {
			binaryBytes, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("failed at read binary content : %w", err)
			}
			return binaryBytes, nil
		}
	}

	return nil, fmt.Errorf("binary file not found at this binary path : %s", binaryPath)
}
