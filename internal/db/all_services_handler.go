package db

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"tosdrgo/handlers/cache"
	"tosdrgo/models"
)

func FetchServicesByGrade(grade string, page int, perPage int) ([]models.SearchResult, int, int, error) {
	normalizedGrade := strings.ToUpper(strings.TrimSpace(grade))

	if normalizedGrade != "A" && normalizedGrade != "B" && normalizedGrade != "C" && normalizedGrade != "D" && normalizedGrade != "E" {
		return nil, 0, http.StatusBadRequest, fmt.Errorf("failed to fetch services")
	}

	total := 0
	for _, service := range allServices {
		if service.Rating == nil || !service.ComprehensivelyReviewed || *service.Rating != normalizedGrade {
			continue
		}
		total++
	}

	cacheKey := fmt.Sprintf("%s_page_%d", normalizedGrade, page)
	if cachedResults, found := cache.GetGradedServices(cacheKey); found {
		return cachedResults, total, http.StatusNotModified, nil
	}

	startIndex := (page - 1) * perPage
	endIndex := startIndex + perPage

	results := make([]models.SearchResult, 0, perPage)
	currentIndex := 0

	for _, service := range allServices {
		if service.Rating == nil || !service.ComprehensivelyReviewed || *service.Rating != normalizedGrade {
			continue
		}

		if currentIndex < startIndex {
			currentIndex++
			continue
		}

		if currentIndex >= endIndex {
			break
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

		result.MatchPercentage = 100
		results = append(results, result)
		currentIndex++
	}

	cache.SetGradedServices(cacheKey, results)
	return results, total, http.StatusOK, nil
}
