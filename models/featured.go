package models

import "time"

type FeaturedService struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Icon   string  `json:"icon"`
	Grade  string  `json:"grade"`
	Points []Point `json:"points"`
}

type FeaturedServices struct {
	Services  []FeaturedService `json:"services"`
	CachedAt  time.Time         `json:"cached_at"`
	ExpiresAt time.Time         `json:"expires_at"`
}
