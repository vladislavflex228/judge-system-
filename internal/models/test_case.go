package models

import "time"

type TestCase struct {
	ID          int64     `json:"id"`
	TaskID      int64     `json:"task_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	InputData   string    `json:"input_data"`
	OutputData  string    `json:"output_data"`
	IsHidden    bool      `json:"is_hidden"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewTestCase(task_id int64, title, description, input_data, output_data string, is_hidden bool, created_at time.Time) *TestCase {
	return &TestCase{TaskID: task_id, Title: title, Description: &description, InputData: input_data, OutputData: output_data, IsHidden: is_hidden, CreatedAt: created_at}
}
