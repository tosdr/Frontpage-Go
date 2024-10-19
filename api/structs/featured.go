package structs

import "time"

type FeaturedResponse struct {
	Data    FeaturedData `json:"data"`
	Message string       `json:"message"`
	Status  string       `json:"status"`
}

type FeaturedData struct {
	FeaturedServices FeaturedServices `json:"featured_services"`
}

type FeaturedServices struct {
	Services  []FeaturedService `json:"services"`
	CachedAt  time.Time         `json:"cached_at"`
	ExpiresAt time.Time         `json:"expires_at"`
}

type FeaturedService struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Icon   string  `json:"icon"`
	Grade  string  `json:"grade"`
	Points []Point `json:"points"`
}
