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
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	FeaturedServices []int `yaml:"featured_services"`
}

var AppConfig *Config

func LoadConfig() error {
	file, err := os.ReadFile("settings.yaml")
	if err != nil {
		return err
	}

	return yaml.Unmarshal(file, &AppConfig)
}
