package models

import "time"

type Task struct {
	ID          int64     `json:id`
	Title       string    `json:title`
	Description string    `json:description`
	LanguageId  int       `json:language_id`
	TimeLimit   int       `json:time_limit_ms`
	MemoryLimit int       `json:memory_limit_mb`
	CreatedAt   time.Time `created_at`
}

func NewTask(language_id, time_limit_ms, memory_limit_mb int, title, description string, created_at time.Time) *Task {
	return &Task{Title: title, Description: description, LanguageId: language_id, TimeLimit: time_limit_ms, MemoryLimit: memory_limit_mb, CreatedAt: created_at}
}
