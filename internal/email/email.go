package email

import (
	"fmt"
	"tosdrgo/internal/config"
	"tosdrgo/internal/logger"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Client struct {
	apiKey string
	from   string
}

var defaultClient *Client

// Init creates a new email client using app config values
func Init() error {
	cfg := config.AppConfig.SMTP
	if cfg.APIKey == "" || cfg.From == "" {
		return fmt.Errorf("missing required email configuration")
	}

	defaultClient = &Client{
		apiKey: cfg.APIKey,
		from:   cfg.From,
	}

	return nil
}

// SendEmail sends an email using the default client
func SendEmail(body string, to string) error {
	if to == "" { // no email
		logger.LogDebug("There is no email.")
		return nil
	}
	if defaultClient == nil {
		return fmt.Errorf("email client not initialized")
	}
	return defaultClient.Send(body, to)
}

// Send sends an email using the client's configuration
func (c *Client) Send(body string, to string) error {
	from := mail.NewEmail("ToS;DR", c.from)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, "ToS;DR Notification", toEmail, "", body)

	client := sendgrid.NewSendClient(c.apiKey)
	_, err := client.Send(message)
	return err
}
