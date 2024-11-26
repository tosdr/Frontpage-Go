package db

import (
	"errors"
	"sort"
	"strings"

	"tosdrgo/cache"
	"tosdrgo/models"
)

func SearchServices(term string) ([]models.SearchResult, error) {
	if len(term) < 3 {
		return nil, errors.New("search term must be at least 3 characters long")
	}

	if cachedResults, found := cache.GetSearchResults(term); found {
		return cachedResults, nil
	}

	query := `
		SELECT id, name, slug, rating, is_comprehensively_reviewed, url
		FROM services
		WHERE LOWER(name) LIKE $1 OR url LIKE $1
	`

	rows, err := DB.Query(query, "%"+strings.ToLower(term)+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		var result models.SearchResult
		var urls string
		err := rows.Scan(&result.ID, &result.Name, &result.Slug, &result.Rating, &result.ComprehensivelyReviewed, &urls)
		if err != nil {
			return nil, err
		}
		result.Image = "https://s3.tosdr.org/logos/" + result.ID + ".png"

		// Calculate match percentage
		nameMatch := calculateSimilarity(strings.ToLower(term), strings.ToLower(result.Name))
		urlMatches := calculateURLMatches(term, strings.Split(urls, ","))
		result.MatchPercentage = max(nameMatch, urlMatches)

		results = append(results, result)
	}

	// Sort results by match percentage
	sort.Slice(results, func(i, j int) bool {
		return results[i].MatchPercentage > results[j].MatchPercentage
	})

	// Return top 20 results
	if len(results) > 20 {
		results = results[:20]
	}

	cache.SetSearchResults(term, results)

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
