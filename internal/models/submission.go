package models

import "time"

type Submission struct {
	ID            int64     `json:"id"`
	TaskID        int64     `json:"task_id"`
	UserID        int64     `json:"user_id"`
	LanguageID    int       `json:"language_id"`
	Code          string    `json:"code"`
	Status        string    `json:"status"`
	ExecutionTime int       `json:"execution_time_ms"`
	MemoryUsed    int       `json:"memory_used_mb"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewSubmission(task_id, user_id int64, language_id, execution_time_ms, memory_used_mb int, code, status string, created_at time.Time) *Submission {
	return &Submission{TaskID: task_id, UserID: user_id, LanguageID: language_id, Code: code, Status: status, ExecutionTime: execution_time_ms, MemoryUsed: memory_used_mb, CreatedAt: created_at}
}
