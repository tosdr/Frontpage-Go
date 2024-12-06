package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"tosdrgo/internal/email"
	"tosdrgo/internal/logger"
)

type ServiceSubmission struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Domains   string `json:"domains"`
	Documents string `json:"documents"`
	Wikipedia string `json:"status"`
	Email     string `json:"email"`
	Note      string `json:"note"`
}

func AddSubmission(submission *ServiceSubmission) error {
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
		submission.Domains,
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

func GetSubmissions(page, perPage int) ([]ServiceSubmission, int, error) {
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

	var submissions []ServiceSubmission
	for rows.Next() {
		var s ServiceSubmission
		err := rows.Scan(&s.ID, &s.Name, &s.Domains, &s.Documents, &s.Wikipedia, &s.Note)
		if err != nil {
			return nil, 0, err
		}
		submissions = append(submissions, s)
	}

	return submissions, total, nil
}

func DeleteSubmission(id string) (*ServiceSubmission, error) {
	logger.LogDebug("Deleting submission with ID: %s", id)

	// First get the full submission
	var submission ServiceSubmission
	err := SubDB.QueryRow(`
		SELECT id, name, domains, documents, wikipedia, email, note 
		FROM service_requests 
		WHERE id = $1`, id).Scan(
		&submission.ID,
		&submission.Name,
		&submission.Domains,
		&submission.Documents,
		&submission.Wikipedia,
		&submission.Email,
		&submission.Note,
	)
	if err != nil {
		logger.LogError(err, "Failed to get submission details")
		return nil, err
	}

	// Prepare delete statement
	stmt, err := SubDB.Prepare(`
		DELETE FROM service_requests 
		WHERE id = $1
	`)
	if err != nil {
		logger.LogError(err, "Failed to prepare delete statement")
		return nil, err
	}
	defer stmt.Close()

	// Execute the delete
	result, err := stmt.Exec(id)
	if err != nil {
		logger.LogError(err, "Failed to execute delete statement")
		return nil, err
	}

	// Check if any row was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.LogError(err, "Failed to get rows affected")
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("no submission found with ID %s", id)
	}

	logger.LogDebug("Successfully deleted submission %s", id)
	return &submission, nil
}

func UpdateSubmissionStatus(id string, action string) error {
	logger.LogDebug("Updating submission status - ID: %s, Action: %s", id, action)

	if action == "deny" {
		service, err := DeleteSubmission(id)
		if err != nil {
			logger.LogError(err, "Failed to delete submission")
			return err
		}

		// Parse and execute the email template
		tmpl, err := template.ParseFiles("templates/emails/denied.gohtml")
		if err != nil {
			logger.LogError(err, "Failed to parse email template")
			return err
		}

		var emailBody bytes.Buffer
		err = tmpl.ExecuteTemplate(&emailBody, "email", struct {
			ServiceName string
			ServicePage string
		}{
			ServiceName: service.Name,
			ServicePage: service.Domains,
		})
		if err != nil {
			logger.LogError(err, "Failed to execute email template")
			return err
		}

		err = email.SendEmail(service.Email, "ToS;DR Service Submission Update", emailBody.String())
		if err != nil {
			logger.LogError(err, "Failed to send email")
		}
		return nil
	} else if action == "accept" {
		logger.LogDebug("Accepted submission %s", id)
		// TODO: Implement accept logic
		return nil
	}

	return fmt.Errorf("invalid action: %s", action)
}
