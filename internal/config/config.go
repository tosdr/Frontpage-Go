package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`
	SubmissionsDatabase struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"submission_database"`
	SMTP struct {
		APIKey string `yaml:"api_key"`
		From   string `yaml:"from"`
	} `yaml:"smtp"`
	Webhook          string `yaml:"webhook"`
	FeaturedServices []int  `yaml:"featured_services"`
	MetricsUsername  string `yaml:"metrics_username"`
	MetricsPassword  string `yaml:"metrics_password"`
	Login            struct {
		Domain       string `yaml:"domain"`
		LogoutReturn string `yaml:"logout_return"`
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		RedirectURI  string `yaml:"redirect_uri"`
		SessionKey   string `yaml:"session_key"`
	} `yaml:"login"`
}

var AppConfig *Config

func LoadConfig() error {
	file, err := os.ReadFile("settings.yaml")
	if err != nil {
		return err
	}

	return yaml.Unmarshal(file, &AppConfig)
}
