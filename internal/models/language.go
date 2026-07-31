package models

type Language struct {
	ID       int    `json:id`
	Name     string `json:name`
	Slug     string `json:slug`
	BuildCmd string `json:build_command`
	ExeCmd   string `json:execution_command`
	IsActive bool   `json:is_active`
}

func NewLanguage(name, slug, buildcmd, execmd string, isactive bool) *Language {
	return &Language{Name: name, Slug: slug, BuildCmd: buildcmd, ExeCmd: execmd, IsActive: isactive}
}
