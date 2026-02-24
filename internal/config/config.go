package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	HTTP struct {
		Addr    string `json:"addr"`
		BaseURL string `json:"base_url"`
	} `json:"http"`
	DB struct {
		DSN string `json:"dsn"`
	} `json:"db"`
	Auth struct {
		CookieName      string `json:"cookie_name"`
		CookieSecret    string `json:"cookie_secret"`
		SessionTTLHours int    `json:"session_ttl_hours"`
	} `json:"auth"`
	SES struct {
		Region          string `json:"region"`
		FromEmail       string `json:"from_email"`
		AccessKeyID     string `json:"access_key_id"`
		SecretAccessKey string `json:"secret_access_key"`
	} `json:"ses"`
}

func LoadConfig() (*Config, error) {
	configPath := os.Getenv("APP_CONFIG")
	if configPath == "" {
		configPath = "./config/config.local.json"
	}

	b, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
