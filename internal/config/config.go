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
	HTTPAddr    string

	// CookieSecure marks session cookies Secure (set true behind HTTPS).
	CookieSecure bool

	// StaticDir is the directory of built frontend files to serve (empty
	// disables static serving, e.g. in local dev where Vite serves the UI).
	StaticDir string
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

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}
	// Many hosts (Render, etc.) inject the port to listen on via PORT.
	if port := os.Getenv("PORT"); port != "" {
		httpAddr = ":" + port
	}

	return &Config{
		DatabaseURL:  dbURL,
		HTTPAddr:     httpAddr,
		CookieSecure: os.Getenv("COOKIE_SECURE") == "true",
		StaticDir:    os.Getenv("STATIC_DIR"),
	}, nil
}
