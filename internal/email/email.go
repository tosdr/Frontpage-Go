package email

import (
	"fmt"
	"tosdrgo/internal/config"
	"tosdrgo/internal/logger"

	"gopkg.in/mail.v2"
)

type Client struct {
	host     string
	port     int
	username string
	password string
	from     string
}

var defaultClient *Client

// Init creates a new email client using app config values
func Init() error {
	cfg := config.AppConfig.SMTP
	if cfg.Host == "" || cfg.User == "" || cfg.Password == "" {
		return fmt.Errorf("missing required email configuration")
	}

	defaultClient = &Client{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.User,
		password: cfg.Password,
		from:     cfg.From,
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
	m := mail.NewMessage()
	m.SetHeader("From", c.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "ToS;DR Notification")
	m.SetBody("text/plain", body)

	d := mail.NewDialer(c.host, c.port, c.username, c.password)
	return d.DialAndSend(m)
}
