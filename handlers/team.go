package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
	"tosdrgo/models"

	"github.com/patrickmn/go-cache"
)

var teamCache = cache.New(24*time.Hour, 24*time.Hour)

func HandleTeamAction(w http.ResponseWriter, r *http.Request) {
	const cacheKey = "team_data"

	// Check cache first
	if cachedData, found := teamCache.Get(cacheKey); found {
		w.Header().Set(ContentType, ContentTypeJson)
		json.NewEncoder(w).Encode(cachedData)
		return
	}

	// Read the JSON file
	jsonData, err := os.ReadFile("assets/about.json")
	if err != nil {
		http.Error(w, "Failed to load team data", http.StatusInternalServerError)
		return
	}

	var team models.Team
	if err := json.Unmarshal(jsonData, &team); err != nil {
		http.Error(w, "Failed to parse team data", http.StatusInternalServerError)
		return
	}

	// Cache the team data
	teamCache.Set(cacheKey, team, cache.DefaultExpiration)

	// Send response
	w.Header().Set(ContentType, ContentTypeJson)
	json.NewEncoder(w).Encode(team)
}
