package db

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"tosdrgo/internal/email"
	"tosdrgo/internal/logger"

	"gorm.io/gorm"
)

type ServiceRequest struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"column:name"`
	Domains   string `gorm:"column:domains"`
	Wikipedia string `gorm:"column:wikipedia"`
	Email     string `gorm:"column:email"`
	Note      string `gorm:"column:note"`
	Count     int    `gorm:"column:count;default:1"`
}

func (ServiceRequest) TableName() string {
	return "service_requests_new"
}

// GetSubmissionsV2 retrieves paginated submissions
func GetSubmissionsV2(page, perPage int) ([]ServiceRequest, int64, error) {
	var submissions []ServiceRequest
	var total int64

	// Get total count
	if err := SubDB.Model(&ServiceRequest{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated submissions
	offset := (page - 1) * perPage
	result := SubDB.Order("count DESC, id DESC").Limit(perPage).Offset(offset).Find(&submissions)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return submissions, total, nil
}

// DeleteSubmissionV2 deletes a submission and returns its data
func DeleteSubmissionV2(id string) (*ServiceRequest, error) {
	logger.LogDebug("Deleting submission with ID: %s", id)

	var submission ServiceRequest
	result := SubDB.First(&submission, id)
	if result.Error != nil {
		logger.LogError(result.Error, "Failed to get submission details")
		return nil, result.Error
	}

	result = SubDB.Delete(&submission)
	if result.Error != nil {
		logger.LogError(result.Error, "Failed to delete submission")
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("no submission found with ID %s", id)
	}

	logger.LogDebug("Successfully deleted submission %s", id)
	return &submission, nil
}

// AddServiceV2 adds a new service to the main database
func AddServiceV2(name string, domains []string, wikipedia string) (uint, error) {
	service := struct {
		gorm.Model
		Name      string
		URL       string
		Wikipedia string
	}{
		Name:      name,
		URL:       strings.Join(domains, ", "),
		Wikipedia: wikipedia,
	}

	_, err := DB.Exec(`
		INSERT INTO services (name, url, wikipedia, created_at, updated_at) 
		VALUES (?, ?, ?, NOW(), NOW()) 
		RETURNING id`, service.Name, service.URL, service.Wikipedia)

	if err != nil {
		logger.LogError(err, "Failed to insert service")
		return 0, err
	}

	return service.ID, nil
}

// AcceptSubmissionV2 accepts a submission and returns the service name, email, and newly created service ID
func AcceptSubmissionV2(id string) (string, string, uint, error) {
	logger.LogDebug("Accepting submission %s", id)

	submission, err := DeleteSubmissionV2(id)
	if err != nil {
		logger.LogError(err, "Failed to delete submission")
		return "", "", 0, err
	}

	// Add service to DB
	serviceID, err := AddServiceV2(submission.Name, strings.Split(submission.Domains, ","), submission.Wikipedia)
	if err != nil {
		logger.LogError(err, "Failed to add service to DB")
		return "", "", 0, err
	}

	return submission.Name, submission.Email, serviceID, nil
}

// DenyRequestV2 denies a submission request and sends notification email
func DenyRequestV2(id string) error {
	submission, err := DeleteSubmissionV2(id)
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
		ServiceName: submission.Name,
		ServicePage: submission.Domains,
	})
	if err != nil {
		logger.LogError(err, "Failed to execute email template")
		return err
	}

	err = email.SendEmail(submission.Email, "ToS;DR Service Submission Update", emailBody.String())
	if err != nil {
		logger.LogError(err, "Failed to send email")
	}

	return nil
}

// AllowRequestV2 accepts a submission request and sends notification email
func AllowRequestV2(id string) error {
	serviceName, serviceEmail, serviceID, err := AcceptSubmissionV2(id)
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
		ServiceID   uint
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

// UpdateSubmissionStatusV2 updates the status of a submission
func UpdateSubmissionStatusV2(id string, action string) error {
	switch action {
	case "allow":
		return AllowRequestV2(id)
	case "deny":
		return DenyRequestV2(id)
	default:
		return fmt.Errorf("invalid action: %s", action)
	}
}

// GetServiceSubmissionByDomainV2 finds a submission by domain
func GetServiceSubmissionByDomainV2(domain string) (uint, error) {
	var submission ServiceRequest
	result := SubDB.Where("? = ANY(domains)", domain).First(&submission)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, result.Error
	}
	return submission.ID, nil
}

// BumpServiceSubmissionCountV2 increments the submission count
func BumpServiceSubmissionCountV2(submissionID uint) error {
	result := SubDB.Model(&ServiceRequest{}).
		Where("id = ?", submissionID).
		Update("count", gorm.Expr("count + ?", 1))
	return result.Error
}

// SearchSubmissionsV2 searches for submissions
func SearchSubmissionsV2(term string, page, perPage int) ([]ServiceRequest, int64, error) {
	var submissions []ServiceRequest
	var total int64
	offset := (page - 1) * perPage

	// Build search condition
	searchCondition := SubDB.Where(
		"name ILIKE ? OR domains ILIKE ?",
		"%"+term+"%",
		"%"+term+"%",
	)

	// Get total count
	if err := searchCondition.Model(&ServiceRequest{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	result := searchCondition.
		Order("count DESC, id DESC").
		Limit(perPage).
		Offset(offset).
		Find(&submissions)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return submissions, total, nil
}
