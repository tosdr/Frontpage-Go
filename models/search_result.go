package models

type SearchResult struct {
	ID                      string  `json:"id"`
	Name                    string  `json:"name"`
	Slug                    string  `json:"slug"`
	ComprehensivelyReviewed bool    `json:"comprehensively_reviewed"`
	Rating                  string  `json:"rating"`
	Image                   string  `json:"image"`
	MatchPercentage         float64 `json:"match_percentage"`
}

type SearchService struct {
	ID                      int
	Name                    string
	URL                     string
	ComprehensivelyReviewed bool
	Rating                  *string
}
