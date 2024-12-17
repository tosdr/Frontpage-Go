package db

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"tosdrgo/handlers/cache"
	"tosdrgo/internal/logger"
	"tosdrgo/models"
)

const (
	maxResults   = 20
	minSearchLen = 3
)

var allServices []models.SearchService

func IndexSearch() {
	start := time.Now()
	logger.LogDebug("Running search index...")
	query := `
		SELECT id, name, url, is_comprehensively_reviewed, rating
		FROM services
	`

	rows, err := DB.Query(query)
	if err != nil {
		logger.LogError(err, "Error executing search query")
		return
	}
	defer rows.Close()

	allServices = make([]models.SearchService, 0)
	for rows.Next() {
		var service models.SearchService
		err := rows.Scan(&service.ID, &service.Name, &service.URL, &service.ComprehensivelyReviewed, &service.Rating)
		if err != nil {
			logger.LogError(err, "Error scanning search query results")
			return
		}
		allServices = append(allServices, service)
	}

	logger.LogDebug("Search index completed in %s", time.Since(start))
}

func SearchServices(term string, grade string) ([]models.SearchResult, int, error) {
	normalizedTerm := strings.ToLower(strings.TrimSpace(term))
	normalizedGrade := strings.ToUpper(strings.TrimSpace(grade))

	if normalizedTerm == "x" {
		normalizedTerm = "twitter"
	}

	if len(normalizedTerm) < minSearchLen {
		logger.LogDebug("Search term too short: %s", normalizedTerm)
		return nil, http.StatusBadRequest, fmt.Errorf("search term must be at least %d characters long", minSearchLen)
	}

	if cachedResults, found := cache.GetSearchResults(normalizedTerm, normalizedGrade); found {
		return cachedResults, http.StatusNotModified, nil
	}

	results := make([]models.SearchResult, 0, 20)
	for _, service := range allServices {
		if !strings.Contains(strings.ToLower(service.Name), normalizedTerm) &&
			!strings.Contains(strings.ToLower(service.URL), normalizedTerm) {
			continue
		}

		if normalizedGrade != "" &&
			(service.Rating == nil || !service.ComprehensivelyReviewed || *service.Rating != normalizedGrade) {
			continue
		}

		var rating string
		if service.Rating != nil && service.ComprehensivelyReviewed {
			rating = *service.Rating
		} else {
			rating = "N/A"
		}

		result := models.SearchResult{
			ID:                      strconv.Itoa(service.ID),
			Name:                    service.Name,
			ComprehensivelyReviewed: service.ComprehensivelyReviewed,
			Rating:                  rating,
			Image:                   "https://s3.tosdr.org/logos/" + strconv.Itoa(service.ID) + ".png",
		}

		nameMatch := calculateQuickSimilarity(normalizedTerm, strings.ToLower(service.Name))
		urlMatch := calculateQuickSimilarity(normalizedTerm, strings.ToLower(service.URL))
		result.MatchPercentage = max(nameMatch, urlMatch)

		if result.MatchPercentage > 0 {
			results = append(results, result)
		}
	}

	if len(results) > maxResults {
		partialSort(results, maxResults)
		results = results[:maxResults]
	} else {
		sort.Slice(results, func(i, j int) bool {
			return results[i].MatchPercentage > results[j].MatchPercentage
		})
	}

	cache.SetSearchResults(normalizedTerm, normalizedGrade, results)
	return results, http.StatusOK, nil
}

func calculateQuickSimilarity(term, target string) float64 {
	if term == target {
		return 100
	}

	if strings.Contains(target, term) {
		return 90 * (float64(len(term)) / float64(len(target)))
	}

	matches := 0
	termIndex := 0
	for i := 0; i < len(target) && termIndex < len(term); i++ {
		if target[i] == term[termIndex] {
			matches++
			termIndex++
		}
	}

	return 100 * float64(matches) / max(float64(len(term)), float64(len(target)))
}

func partialSort(results []models.SearchResult, k int) {
	for i := 0; i < k; i++ {
		maxIdx := i
		for j := i + 1; j < len(results); j++ {
			if results[j].MatchPercentage > results[maxIdx].MatchPercentage {
				maxIdx = j
			}
		}
		results[i], results[maxIdx] = results[maxIdx], results[i]
	}
}
