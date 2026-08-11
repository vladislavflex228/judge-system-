package domain

import "time"

type Verdict string

const (
	VerdictOK  Verdict = "OK"  // Accepted
	VerdictWA  Verdict = "WA"  // Wrong Answer
	VerdictTLE Verdict = "TLE" // Time Limit Exceeded
	VerdictMLE Verdict = "MLE" // Memory Limit Exceeded
	VerdictRE  Verdict = "RE"  // Runtime Error
	VerdictCE  Verdict = "CE"  // Compilation Error
	VerdictSE  Verdict = "SE"  // System Error
)

type SingleTest struct {
	ID             int64
	Input          string
	ExpectedOutput string
}

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

type TestResult struct {
	TestID   int64
	Verdict  Verdict
	TimeSec  float64
	MemoryKB int64
	ErrorLog string
}
