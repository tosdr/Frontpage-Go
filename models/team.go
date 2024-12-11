package models

import (
	"html/template"
)

type TeamMember struct {
	Photo       string            `json:"photo"`
	Name        string            `json:"name"`
	Title       string            `json:"title"`
	Description template.HTML     `json:"description"`
	Links       map[string]string `json:"links"`
}

type Team struct {
	Founders []TeamMember `json:"founders"`
	Current  []TeamMember `json:"current"`
	Past     []TeamMember `json:"past"`
}
