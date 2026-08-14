package domain

import "time"

type ExecutionResult struct {
	ExitCode   int
	Stdout     string
	Stderr     string
	Duration   time.Duration
	MemoryUsed int64
	MemoryKB   int64
	MLE        bool
	TLE        bool
	Binary     []byte
}
