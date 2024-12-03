package db

import (
	"errors"
	"sort"
	"strings"
	"tosdrgo/logger"

	"tosdrgo/cache"
	"tosdrgo/models"
)

const (
	maxResults   = 20
	minSearchLen = 3
)

func SearchServices(term string) ([]models.SearchResult, error) {
	normalizedTerm := strings.ToLower(strings.TrimSpace(term))

	if normalizedTerm == "x" { // twitter is... special. thanks elon.
		normalizedTerm = "twitter"
	}

	if len(normalizedTerm) < minSearchLen {
		return nil, errors.New("search term must be at least 3 characters long")
	}

	if cachedResults, found := cache.GetSearchResults(normalizedTerm); found {
		return cachedResults, nil
	}

	query := `
		SELECT id, name, slug, rating, is_comprehensively_reviewed, url
		FROM services
		WHERE LOWER(name) LIKE $1 OR url LIKE $1
		LIMIT 100
	`

	if strings.Count(normalizedTerm, "%") > 2 {
		logger.LogWarn("Potentially expensive search query with multiple wildcards: %s", normalizedTerm)
	}

	rows, err := DB.Query(query, "%"+normalizedTerm+"%")
	if err != nil {
		logger.LogError(err, "Error executing search query")
		return nil, err
	}
	defer rows.Close()

	results := make([]models.SearchResult, 0, 100)
	for rows.Next() {
		var result models.SearchResult
		var urls string
		err := rows.Scan(&result.ID, &result.Name, &result.Slug, &result.Rating, &result.ComprehensivelyReviewed, &urls)
		if err != nil {
			logger.LogError(err, "Error scanning search query results")
			return nil, err
		}
		result.Image = "https://s3.tosdr.org/logos/" + result.ID + ".png"

		urlList := strings.Split(urls, ",")

		nameMatch := calculateSimilarity(normalizedTerm, strings.ToLower(result.Name))
		urlMatches := calculateURLMatches(normalizedTerm, urlList)
		result.MatchPercentage = max(nameMatch, urlMatches)

		results = append(results, result)
	}

	if len(results) > maxResults {
		partialSort(results, maxResults)
		results = results[:maxResults]
	} else {
		sort.Slice(results, func(i, j int) bool {
			return results[i].MatchPercentage > results[j].MatchPercentage
		})
	}

	cache.SetSearchResults(normalizedTerm, results)
	return results, nil
}

func calculateURLMatches(term string, urls []string) float64 {
	maxMatch := 0.0
	for _, url := range urls {
		match := calculateSimilarity(strings.ToLower(term), strings.ToLower(url))
		if match > maxMatch {
			maxMatch = match
		}
	}
	return maxMatch
}

func calculateSimilarity(s1, s2 string) float64 {
	distance := levenshteinDistance(s1, s2)
	maxLen := max(float64(len(s1)), float64(len(s2)))
	return 100 * (1 - float64(distance)/maxLen)
}

func levenshteinDistance(s1, s2 string) int {
	m := len(s1)
	n := len(s2)
	d := make([][]int, m+1)
	for i := range d {
		d[i] = make([]int, n+1)
	}

	for i := 0; i <= m; i++ {
		d[i][0] = i
	}
	for j := 0; j <= n; j++ {
		d[0][j] = j
	}

	for j := 1; j <= n; j++ {
		for i := 1; i <= m; i++ {
			if s1[i-1] == s2[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				d[i][j] = min(d[i-1][j]+1, min(d[i][j-1]+1, d[i-1][j-1]+1))
			}
		}
	}
	return d[m][n]
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
