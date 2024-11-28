package db

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"tosdrgo/cache"
	"tosdrgo/logger"
	"tosdrgo/models"
)

func FetchServiceData(serviceID int) (*models.Service, error) {
	if cachedService, found := cache.GetService(serviceID); found {
		return cachedService, nil
	}

	service, err := fetchBaseServiceData(serviceID)
	if err != nil {
		logger.LogError(err, fmt.Sprintf("Error fetching base service data for ID %d", serviceID))
		return nil, err
	}

	if err := fetchRelatedData(service); err != nil {
		return nil, err
	}

	service.Points = organizePoints(service.Points)
	service.Image = "https://s3.tosdr.org/logos/" + strconv.Itoa(service.ID) + ".png"

	cache.SetService(serviceID, service)
	return service, nil
}

func fetchBaseServiceData(serviceID int) (*models.Service, error) {
	var service models.Service
	var urlString string

	err := DB.QueryRow(`
		SELECT id, name, updated_at, created_at, slug, rating, is_comprehensively_reviewed, url 
			FROM services WHERE id = $1
	`, serviceID).Scan(
		&service.ID, &service.Name, &service.UpdatedAt, &service.CreatedAt,
		&service.Slug, &service.Rating, &service.ComprehensivelyReviewed, &urlString,
	)
	if err != nil {
		logger.LogError(err, fmt.Sprintf("Error fetching base service data for ID %d", serviceID))
		return nil, err
	}

	if urlString != "" {
		service.URLs = strings.Split(urlString, ",")
	}
	return &service, nil
}

func fetchRelatedData(service *models.Service) error {
	docChan := make(chan []models.Document, 1)
	pointsChan := make(chan []models.Point, 1)
	errChan := make(chan error, 2)

	go fetchDocuments(service.ID, docChan, errChan)
	go fetchPoints(service.ID, pointsChan, errChan)

	var err1, err2 error
	service.Documents = <-docChan
	service.Points = <-pointsChan
	err1 = <-errChan
	err2 = <-errChan

	if err1 != nil || err2 != nil {
		logger.LogError(fmt.Errorf("error fetching data: %v, %v", err1, err2), fmt.Sprintf("Error fetching related data for service ID %d", service.ID))
		return fmt.Errorf("error fetching data: %v, %v", err1, err2)
	}
	return nil
}

func fetchDocuments(serviceID int, docChan chan<- []models.Document, errChan chan<- error) {
	rows, err := DB.Query(`
		SELECT id, name, url, created_at, updated_at 
		FROM documents WHERE service_id = $1
	`, serviceID)
	if err != nil {
		logger.LogError(err, fmt.Sprintf("Error fetching documents for service ID %d", serviceID))
		errChan <- err
		return
	}
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var doc models.Document
		if err := rows.Scan(&doc.ID, &doc.Name, &doc.URL, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			errChan <- err
			return
		}
		docs = append(docs, doc)
	}
	docChan <- docs
	errChan <- nil
}

func fetchPoints(serviceID int, pointsChan chan<- []models.Point, errChan chan<- error) {
	rows, err := DB.Query(`
		SELECT 
			p.id, p.title, p.source, p.status, p.analysis, p.document_id, p.updated_at, p.created_at,
			c.id, c.score, c.title, c.description, c.updated_at, c.created_at, c.topic_id, c.classification
		FROM points p
		LEFT JOIN cases c ON p.case_id = c.id
		WHERE p.service_id = $1 AND p.status = 'approved'
		ORDER BY c.score DESC
	`, serviceID)
	if err != nil {
		logger.LogError(err, fmt.Sprintf("Error fetching points for service ID %d", serviceID))
		errChan <- err
		return
	}
	defer rows.Close()

	points := scanPoints(rows)
	pointsChan <- points
	errChan <- nil
}

func scanPoints(rows *sql.Rows) []models.Point {
	var points []models.Point
	for rows.Next() {
		var point models.Point
		var caseData models.Case
		var documentID, caseID sql.NullInt64
		if err := rows.Scan(
			&point.ID, &point.Title, &point.Source, &point.Status, &point.Analysis,
			&documentID, &point.UpdatedAt, &point.CreatedAt,
			&caseID, &caseData.Weight, &caseData.Title, &caseData.Description,
			&caseData.UpdatedAt, &caseData.CreatedAt, &caseData.TopicID, &caseData.Classification,
		); err != nil {
			continue
		}
		if documentID.Valid {
			point.DocumentID = int(documentID.Int64)
		}
		if caseID.Valid {
			caseData.ID = int(caseID.Int64)
			point.Case = &caseData
		}
		points = append(points, point)
	}
	return points
}

func organizePoints(points []models.Point) []models.Point {
	classificationGroups := groupPointsByClassification(points)
	sortedPoints := sortPointsByClassification(classificationGroups)
	return appendPointsWithoutCase(sortedPoints, points)
}

func groupPointsByClassification(points []models.Point) map[string][]models.Point {
	groups := map[string][]models.Point{
		"blocker": {}, "bad": {}, "good": {}, "neutral": {},
	}

	for _, point := range points {
		if point.Case != nil {
			groups[point.Case.Classification] = append(groups[point.Case.Classification], point)
		}
	}

	for _, classification := range []string{"blocker", "bad", "good", "neutral"} {
		sort.Slice(groups[classification], func(i, j int) bool {
			return groups[classification][i].Case.Weight > groups[classification][j].Case.Weight
		})
	}
	return groups
}

func sortPointsByClassification(groups map[string][]models.Point) []models.Point {
	var sorted []models.Point
	for _, classification := range []string{"blocker", "bad", "good", "neutral"} {
		sorted = append(sorted, groups[classification]...)
	}
	return sorted
}

func appendPointsWithoutCase(sorted []models.Point, points []models.Point) []models.Point {
	for _, point := range points {
		if point.Case == nil {
			sorted = append(sorted, point)
		}
	}
	return sorted
}
