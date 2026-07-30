package models

type Language struct {
	ID       int    `json:id`
	Name     string `json:name`
	Slug     string `json:slug`
	BuildCmd string `json:build_command`
	ExeCmd   string `json:execution_command`
	IsActive bool   `json:is_active`
}
