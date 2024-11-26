package db

import (
	"database/sql"
	"sort"
	"strconv"
	"strings"

	"tosdrgo/cache"
	"tosdrgo/models"

	"log"
)

func FetchServiceData(serviceID int) (*models.Service, error) {
	if cachedService, found := cache.GetService(serviceID); found {
		return cachedService, nil
	}

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
		log.Printf("Error fetching service data for ID %d: %v", serviceID, err)
		return nil, err
	}

	// Split the comma-separated URL string into a slice
	if urlString != "" {
		service.URLs = strings.Split(urlString, ",")
	}

	// Fetch documents
	rows, err := DB.Query(`
		SELECT id, name, url, created_at, updated_at 
		FROM documents WHERE service_id = $1
	`, serviceID)
	if err != nil {
		log.Printf("Error fetching documents for service ID %d: %v", serviceID, err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var doc models.Document
		if err := rows.Scan(&doc.ID, &doc.Name, &doc.URL, &doc.CreatedAt, &doc.UpdatedAt); err != nil {
			return nil, err
		}
		service.Documents = append(service.Documents, doc)
	}

	// Fetch points and cases
	rows, err = DB.Query(`
		SELECT 
			p.id, p.title, p.source, p.status, p.analysis, p.document_id, p.updated_at, p.created_at,
			c.id, c.score, c.title, c.description, c.updated_at, c.created_at, c.topic_id, c.classification
		FROM points p
		LEFT JOIN cases c ON p.case_id = c.id
		WHERE p.service_id = $1 AND p.status = 'approved'
		ORDER BY c.score DESC
	`, serviceID)
	if err != nil {
		log.Printf("Error fetching points and cases for service ID %d: %v", serviceID, err)
		return nil, err
	}
	defer rows.Close()

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
			return nil, err
		}
		if documentID.Valid {
			point.DocumentID = int(documentID.Int64)
		}
		if caseID.Valid {
			caseData.ID = int(caseID.Int64)
			point.Case = &caseData
		}
		service.Points = append(service.Points, point)
	}

	// Group points by classification
	classificationGroups := map[string][]models.Point{
		"blocker": {},
		"bad":     {},
		"good":    {},
		"neutral": {},
	}

	for _, point := range service.Points {
		if point.Case != nil {
			classificationGroups[point.Case.Classification] = append(classificationGroups[point.Case.Classification], point)
		}
	}

	// Sort points within each classification group
	for _, classification := range []string{"blocker", "bad", "good", "neutral"} {
		sort.Slice(classificationGroups[classification], func(i, j int) bool {
			return classificationGroups[classification][i].Case.Weight > classificationGroups[classification][j].Case.Weight
		})
	}

	// Combine sorted groups back into service.Points
	service.Points = []models.Point{}
	for _, classification := range []string{"blocker", "bad", "good", "neutral"} {
		service.Points = append(service.Points, classificationGroups[classification]...)
	}

	// Add points without cases at the end
	for _, point := range service.Points {
		if point.Case == nil {
			service.Points = append(service.Points, point)
		}
	}

	// Set default values
	service.Image = "https://s3.tosdr.org/logos/" + strconv.Itoa(service.ID) + ".png"

	cache.SetService(serviceID, &service)

	return &service, nil
}
