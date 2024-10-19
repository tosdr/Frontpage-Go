package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	APIBaseURL string `yaml:"api_base_url"`
	ServerPort int    `yaml:"server_port"`
}

var AppConfig Config

func LoadConfig() error {
	file, err := os.ReadFile("settings.yaml")
	if err != nil {
		return err
	}

	return yaml.Unmarshal(file, &AppConfig)
}
