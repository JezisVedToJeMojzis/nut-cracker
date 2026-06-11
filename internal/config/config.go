// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the application.
type Config struct {
	DatabaseURL string
}

// Load reads configuration from the environment. In development it first
// attempts to load a .env file (ignored if absent, e.g. in production where
// real environment variables are set instead).
func Load() (*Config, error) {
	// Best-effort: a missing .env is not an error (production uses real env vars).
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	return &Config{
		DatabaseURL: dbURL,
	}, nil
}
