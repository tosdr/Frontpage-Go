package models

import (
	"time"
)

type ServiceSubmission struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Domains   []string   `json:"domains"`
	Documents []Document `json:"documents"`
	Wikipedia string     `json:"wikipedia"`
	Email     string     `json:"email"`
	Note      string     `json:"note"`
	CreatedAt time.Time  `json:"created_at"`
	Status    string     `json:"status"`
}

type SubmissionDocument struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
