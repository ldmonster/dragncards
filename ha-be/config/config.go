package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		URL string `yaml:"url"`
	} `yaml:"database"`
	Log struct {
		Level string `yaml:"level"`
	} `yaml:"log"`
	Redis struct {
		URL string `yaml:"url"`
	} `yaml:"redis"`
	Auth struct {
		JWTSecret       string `yaml:"jwt_secret"`
		AccessLifetime  int    `yaml:"access_lifetime_minutes"`
		RefreshLifetime int    `yaml:"refresh_lifetime_hours"`
	} `yaml:"auth"`
	Email struct {
		SMTPHost     string `yaml:"smtp_host"`
		SMTPPort     int    `yaml:"smtp_port"`
		SMTPUsername string `yaml:"smtp_username"`
		SMTPPassword string `yaml:"smtp_password"`
	} `yaml:"email"`
	Patreon struct {
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
	} `yaml:"patreon"`
	Recaptcha struct {
		SiteKey   string `yaml:"site_key"`
		SecretKey string `yaml:"secret_key"`
	} `yaml:"recaptcha"`
}

func defaultConfig() *Config {
	cfg := &Config{}
	cfg.Server.Port = 9000
	cfg.Database.URL = "postgres://localhost:5432/dragncards?sslmode=disable"
	cfg.Log.Level = "info"
	cfg.Redis.URL = ""
	cfg.Auth.JWTSecret = "default-secret"
	cfg.Auth.AccessLifetime = 30
	cfg.Auth.RefreshLifetime = 90 * 24
	cfg.Email.SMTPHost = ""
	cfg.Email.SMTPPort = 587
	cfg.Email.SMTPUsername = ""
	cfg.Email.SMTPPassword = ""
	cfg.Patreon.ClientID = ""
	cfg.Patreon.ClientSecret = ""
	cfg.Recaptcha.SiteKey = ""
	cfg.Recaptcha.SecretKey = ""
	return cfg
}

func LoadConfig(path string) (*Config, error) {
	cfg := defaultConfig()
	if path == "" {
		return cfg, nil
	}

	cleanPath := strings.TrimSpace(path)
	if cleanPath == "config/config.yaml.example" {
		// use defaults in case example file isn't present
		return cfg, nil
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return cfg, fmt.Errorf("read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("parse config file: %w", err)
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 9000
	}
	if cfg.Database.URL == "" {
		cfg.Database.URL = "postgres://localhost:5432/dragncards?sslmode=disable"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}

	return cfg, nil
}
