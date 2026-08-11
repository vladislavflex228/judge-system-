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
	MaxUsage uint64 `json:"max_usage"` // Пиковое потребление памяти (Cgroups v1)
	Stats    struct {
		Peak     uint64 `json:"peak"` // Пиковое потребление в Cgroups v2
		Anon     uint64 `json:"anon"` // Анонимная память (без cache/buffers)
		Inactive uint64 `json:"inactive_file"`
	} `json:"stats"`
}

type ContainerStatsResponse struct {
	MemoryStats MemoryStats `json:"memory_stats"`
}

func (r *DockerRunner) GetMemoryUsage(ctx context.Context, containerID string) (uint64, error) {
	// Запрашиваем однократную статистику контейнера
	stats, err := r.cli.ContainerStats(ctx, containerID, client.ContainerStatsOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get stats: %w", err)
	}
	defer stats.Body.Close()

	var statsJSON ContainerStatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		return 0, fmt.Errorf("failed to decode stats json: %w", err)
	}

	// В зависимости от версии Cgroups на сервере (v1 или v2)
	maxMemory := statsJSON.MemoryStats.MaxUsage
	if maxMemory == 0 {
		maxMemory = statsJSON.MemoryStats.Stats.Peak
	}

	// Для более точной оценки вычитаем кеш файловой системы, оставляя чистый RSS
	cleanMemory := maxMemory - statsJSON.MemoryStats.Stats.Inactive

	return cleanMemory, nil
}

// createTarArchive упаковывает строку с кодом в tar-архив в оперативной памяти
func createTarArchive(filename string, content []byte) (io.Reader, error) {
	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)

	// Создаем заголовок файла для tar-архива
	header := &tar.Header{
		Name:    filename,            // Имя файла (например, "solution.cpp")
		Size:    int64(len(content)), // Размер файла
		Mode:    0644,                // Права доступа (чтение/запись для владельца, чтение для остальных)
		ModTime: time.Now(),          // Время модификации
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

func (r *DockerRunner) ExtractBinary(ctx context.Context, container_id string, binaryPath string) ([]byte, error) {
	copyRes, err := r.cli.CopyFromContainer(ctx, container_id, client.CopyFromContainerOptions{SourcePath: binaryPath})

	if err != nil {
		return nil, fmt.Errorf("failed at copy from compile container : %w", err)
	}

	defer copyRes.Content.Close()

	tarReader := tar.NewReader(copyRes.Content)

	for {
		header, err := tarReader.Next()
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
