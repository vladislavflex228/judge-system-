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
		DockerImage: "1.26-alpine",
		SourceFile:  "solution.go",
		BinaryFile:  "solution",
		CompileCmd:  []string{"go", "build", "-o", "solution", "solution.go"},
		RunCmd:      []string{"./solution"},
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
		WorkingDir:   b.WorkDir,
		Tty:          false, //true - если нужен терминал внутри контейнера
		AttachStdout: true,
		AttachStderr: true,
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:     512 * 1024 * 1024,
			MemorySwap: 512 * 1024 * 1024,
			NanoCPUs:   2 * 1e9,
		},
		ReadonlyRootfs: true,
		Tmpfs: map[string]string{
			b.WorkDir: "rw,exec,nosuid,size=128m",
			"/tmp":    "rw,noexec,nosuid,size=32m",
		},
		CapDrop: []string{"ALL"},
	}

	return config, hostConfig
}

func (b *ConfigBuilder) BuildRunCmd(image string, runCmd []string, timeLimSec float64, memoryLimMb int64) (*container.Config, *container.HostConfig) {
	config := &container.Config{
		Image:           image,
		Cmd:             runCmd,
		WorkingDir:      b.WorkDir,
		NetworkDisabled: true,
		AttachStdin:     true,
		AttachStdout:    true,
		AttachStderr:    true,
		OpenStdin:       true,
		StdinOnce:       true,
		Tty:             false,
		User:            "1000:1000", //Non-root user
	}

	var a int64 = 32

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:     memoryLimMb * 1024 * 1024,
			MemorySwap: memoryLimMb * 1024 * 1024,
			NanoCPUs:   int64(timeLimSec * 1e9),
			PidsLimit:  &a,
		},
		ReadonlyRootfs: true,
		Tmpfs: map[string]string{
			b.WorkDir: "rw,exec,nosuid,size=64m",
			"/tmp":    "rw,noexec,nosuid,size=16m",
		},
		CapDrop: []string{"ALL"},
	}

	return config, hostConfig
}
