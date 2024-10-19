package structs

import (
	"time"
)

type Service struct {
	ID                      int        `json:"id"`
	ComprehensivelyReviewed bool       `json:"comprehensively_reviewed"`
	Name                    string     `json:"name"`
	UpdatedAt               time.Time  `json:"updated_at"`
	CreatedAt               time.Time  `json:"created_at"`
	Slug                    string     `json:"slug"`
	Rating                  string     `json:"rating"`
	Image                   string     `json:"image"`
	URLs                    []string   `json:"urls"`
	Documents               []Document `json:"documents"`
	Points                  []Point    `json:"points"`
}

type Document struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Point struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Source     string    `json:"source"`
	Status     string    `json:"status"`
	Analysis   string    `json:"analysis"`
	Case       *Case     `json:"case"`
	DocumentID int       `json:"document_id"`
	UpdatedAt  time.Time `json:"updated_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type Case struct {
	ID             int       `json:"id"`
	Weight         int       `json:"weight"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	UpdatedAt      time.Time `json:"updated_at"`
	CreatedAt      time.Time `json:"created_at"`
	TopicID        int       `json:"topic_id"`
	Classification string    `json:"classification"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Service Service `json:"service"`
	} `json:"data"`
}
