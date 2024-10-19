package structs

type SearchResponse struct {
	Data    []SearchResult `json:"data"`
	Message string         `json:"message"`
	Status  string         `json:"status"`
}

type SearchResult struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	Rating          string  `json:"rating"`
	Image           string  `json:"image"`
	MatchPercentage float64 `json:"match_percentage"`
}
