// Package config loads and holds application configuration from environment variables.
package config

import (
	"fmt"
	"os"
)

// Config holds all configuration for the application.
type Config struct {
	DatabaseURL string
	Port        string
	JWTSecret   string
}

// AppConfig is the global configuration instance.
var AppConfig *Config

// Load reads environment variables and populates AppConfig.
// It must be called before any other package tries to use AppConfig.
func Load() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}

	AppConfig = &Config{
		DatabaseURL: dbURL,
		Port:        port,
		JWTSecret:   jwtSecret,
	}

	return nil
}
