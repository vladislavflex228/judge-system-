package config

import (
	"github.com/moby/moby/api/types/container"
)

type LanguageConfig struct {
	Slug        string
	DockerImage string
	SourceFile  string
	BinaryFile  string
	CompileCmd  []string
	RunCmd      []string
}

var Registry = map[string]LanguageConfig{
	"cpp20": {
		Slug:        "cpp20",
		DockerImage: "judge-gcc:13",
		SourceFile:  "solution.cpp",
		BinaryFile:  "solution",
		CompileCmd:  []string{"g++", "-O2", "-std=c++20", "solution.cpp", "-o", "solution"},
		RunCmd:      []string{"./solution"},
	},
	"python3": {
		Slug:        "python3",
		DockerImage: "python:3.11-slim",
		SourceFile:  "solution.py",
		CompileCmd:  []string{"python3", "-m", "py_compile", "solution.py"},
		RunCmd:      []string{"python3", "-u", "solution.py"},
	},
	"golang": {
		Slug:        "golang",
		DockerImage: "golang:1.26-alpine",
		SourceFile:  "solution.go",
		BinaryFile:  "solution",
		CompileCmd:  []string{"sh", "-c", "tar -xf - -C /tmp && go build -o /tmp/solution /tmp/solution.go"},
		RunCmd:      []string{"sh", "-cx", "tar -xf - -C /tmp && chmod +x /tmp/solution && /tmp/solution < /tmp/input.txt"},
	},
}

type ConfigBuilder struct { //Pattern Builder:нет пробрасываний переменных по уровням(создаем на уровне конфигурации),маштабируемость,пространство имен,внедрение зависимостей
	WorkDir string
}

func NewConfigBuilder(workDir string) *ConfigBuilder {
	return &ConfigBuilder{WorkDir: workDir}
}

func (b *ConfigBuilder) BuildCompileCmd(image string, compileCmd []string) (*container.Config, *container.HostConfig) {
	config := &container.Config{
		Image:        image,
		Cmd:          compileCmd,
		WorkingDir:   "/tmp",
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    true,
		Env:          []string{"CGO_ENABLED=0", "GOCACHE=/tmp/.cache"},
		StdinOnce:    true,
		Tty:          false,
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:     512 * 1024 * 1024,
			MemorySwap: 512 * 1024 * 1024,
			NanoCPUs:   2 * 1e9,
		},
		ReadonlyRootfs: false,
		AutoRemove:     false,

		CapDrop: []string{"ALL"},
	}

	return config, hostConfig
}

func (b *ConfigBuilder) BuildRunCmd(image string, runCmd []string, cpuLimSec float64, memoryLimMb int64) (*container.Config, *container.HostConfig) {
	config := &container.Config{
		Image:           image,
		Cmd:             runCmd,
		WorkingDir:      "/tmp",
		NetworkDisabled: true, // У контейнера нет маршрутов наружу и внутри виртуальной сети докера
		AttachStdin:     true,
		AttachStdout:    true,
		AttachStderr:    true,
		OpenStdin:       true,
		Env:             []string{"CGO_ENABLED=0", "GOCACHE=/tmp/.cache"},
		StdinOnce:       true,
		Tty:             false,
		User:            "1000:1000", //Non-root user
	}

	var a int64 = 32

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:     memoryLimMb * 1024 * 1024,
			MemorySwap: memoryLimMb * 1024 * 1024,
			NanoCPUs:   int64(cpuLimSec * 1e9),
			PidsLimit:  &a, //защита от fork-bomb
		},
		ReadonlyRootfs: true,
		Tmpfs: map[string]string{
			"/tmp": "rw,exec,nosuid,size=128m",
		},
		CapDrop: []string{"ALL"},
	}

	return config, hostConfig
}
