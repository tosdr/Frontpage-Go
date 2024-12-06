package db

import (
	"encoding/json"
	"strings"
	"tosdrgo/logger"
	"tosdrgo/models"
)

type ServiceSubmissionWithStatus struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Domains   string `json:"domains"`
	Documents string `json:"documents"`
	Wikipedia string `json:"status"`
	Note      string `json:"note"`
}

func AddSubmission(submission *models.ServiceSubmission) error {
	// Convert domains array to comma-separated string
	domainsString := strings.Join(submission.Domains, ",")

	// Convert documents array to JSON string
	documentsJSON, err := json.Marshal(submission.Documents)
	if err != nil {
		logger.LogError(err, "Failed to marshal documents")
		return err
	}

	// Prepare the SQL statement
	stmt, err := SubDB.Prepare(`
		INSERT INTO service_requests 
		(name, domains, documents, wikipedia, email, note) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		logger.LogError(err, "Failed to prepare submission statement")
		return err
	}
	defer stmt.Close()

	// Execute the statement
	_, err = stmt.Exec(
		submission.Name,
		domainsString,
		string(documentsJSON),
		submission.Wikipedia,
		submission.Email,
		submission.Note,
	)

	if err != nil {
		logger.LogError(err, "Failed to execute submission insert")
		return err
	}

	return nil
}

func GetSubmissions(page, perPage int) ([]ServiceSubmissionWithStatus, int, error) {
	offset := (page - 1) * perPage

	// Get total count
	var total int
	err := SubDB.QueryRow("SELECT COUNT(*) FROM service_requests").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated submissions
	rows, err := SubDB.Query(`
		SELECT id, name, domains, documents, wikipedia, note 
		FROM service_requests 
		ORDER BY id DESC 
		LIMIT $1 OFFSET $2
	`, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var submissions []ServiceSubmissionWithStatus
	for rows.Next() {
		var s ServiceSubmissionWithStatus
		err := rows.Scan(&s.ID, &s.Name, &s.Domains, &s.Documents, &s.Wikipedia, &s.Note)
		if err != nil {
			return nil, 0, err
		}
		submissions = append(submissions, s)
	}

	return submissions, total, nil
}
