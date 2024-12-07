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

type DocumentSubmission struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	XPath string `json:"xpath"`
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

func AddService(name string, url string, wikipedia string, user string) (int, error) {
	var serviceID int
	err := DB.QueryRow(`
		INSERT INTO services (name, url, wikipedia, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id
	`, name, url, wikipedia).Scan(&serviceID)
	if err != nil {
		logger.LogError(err, "Failed to insert service")
		return 0, err
	}
	return serviceID, nil
}

// AcceptSubmission accepts a submission and returns the service name, email, newly created service ID, and error
func AcceptSubmission(id string) (string, string, int, error) {
	logger.LogDebug("Accepting submission %s", id)
	service, err := DeleteSubmission(id)
	if err != nil {
		logger.LogError(err, "Failed to delete submission")
		return "", "", 0, err
	}

	// add service to DB
	serviceID, err := AddService(service.Name, service.Domains, service.Wikipedia, service.Email)
	if err != nil {
		logger.LogError(err, "Failed to add service to DB")
		return "", "", 0, err
	}

	// add documents to DB
	var documents []DocumentSubmission
	err = json.Unmarshal([]byte(service.Documents), &documents)
	if err != nil {
		logger.LogError(err, "Failed to unmarshal documents")
		return "", "", 0, err
	}

	for _, document := range documents {
		_, err := AddDocument(document.Name, document.URL, document.XPath, fmt.Sprint(serviceID), service.Email)
		if err != nil {
			logger.LogError(err, "Failed to add document to DB")
			return "", "", 0, err
		}
	}

	return service.Name, service.Email, serviceID, nil
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
		serviceName, serviceEmail, serviceID, err := AcceptSubmission(id)
		if err != nil {
			logger.LogError(err, "Failed to accept submission")
			return err
		}

		// Parse and execute the email template
		tmpl, err := template.ParseFiles("templates/emails/accepted.gohtml")
		if err != nil {
			logger.LogError(err, "Failed to parse email template")
			return err
		}

		var emailBody bytes.Buffer
		err = tmpl.ExecuteTemplate(&emailBody, "email", struct {
			ServiceName string
			ServiceID   int
		}{
			ServiceName: serviceName,
			ServiceID:   serviceID,
		})
		if err != nil {
			logger.LogError(err, "Failed to execute email template")
			return err
		}

		err = email.SendEmail(serviceEmail, "ToS;DR Service Submission Update", emailBody.String())
		if err != nil {
			logger.LogError(err, "Failed to send email")
		}
		return nil
	}

	return fmt.Errorf("invalid action: %s", action)
}

func AddDocument(name string, url string, xpath string, serviceID string, user string) (int, error) {
	// Check if service exists
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM services WHERE id = $1)", serviceID).Scan(&exists)
	if err != nil {
		logger.LogError(err, "Failed to check if service exists")
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("service with ID %s does not exist", serviceID)
	}

	// Insert document
	var documentID int
	err = DB.QueryRow(`
		INSERT INTO documents (name, url, selector, created_at, updated_at, service_id) 
		VALUES ($1, $2, $3, NOW(), NOW(), $4) 
		RETURNING id
	`, name, url, xpath, serviceID).Scan(&documentID)
	if err != nil {
		logger.LogError(err, "Failed to insert document")
		return 0, err
	}

	// Create version entry
	err = createVersion("Document", fmt.Sprint(documentID), "create", "Created document", user, "")
	if err != nil {
		logger.LogError(err, "Failed to create version entry")
		return documentID, err // Still return documentID since document was created
	}

	return documentID, nil
}

func createVersion(itemType string, itemID string, event string, objectChanges string, whodunnit string, object string) error {
	stmt, err := DB.Prepare(`
		INSERT INTO versions (item_type, item_id, event, created_at, object_changes, whodunnit, object) 
		VALUES ($1, $2, $3, NOW(), $4, $5, $6)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(itemType, itemID, event, objectChanges, whodunnit, object)
	return err
}
