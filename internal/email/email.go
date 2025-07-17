package email

import (
	"crypto/tls"
	"fmt"
	"tosdrgo/internal/config"
	"tosdrgo/internal/logger"

	"gopkg.in/gomail.v2"
)

type Client struct {
	host     string
	port     int
	username string
	password string
	from     string
	tls      bool
	dialer   *gomail.Dialer
}

var defaultClient *Client

func Init() error {
	cfg := config.AppConfig.SMTP
	if cfg.Host == "" || cfg.From == "" {
		return fmt.Errorf("missing required email configuration")
	}

	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	if cfg.TLS {
		d.TLSConfig = nil
	} else {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		d.SSL = false
	}

	defaultClient = &Client{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
		tls:      cfg.TLS,
		dialer:   d,
	}

	return nil
}

func SendEmail(to string, subject string, body string) error {
	if to == "" { // no email
		logger.LogDebug("There is no email.")
		return nil
	}
	if defaultClient == nil {
		return fmt.Errorf("email client not initialized")
	}
	return defaultClient.Send(to, subject, body)
}

func (c *Client) Send(to string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", c.from, "ToS;DR Notifications")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	return c.dialer.DialAndSend(m)
}
