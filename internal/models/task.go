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
