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

	// AppBaseURL is the public URL of the frontend, used to build links in
	// emails (e.g. verification / password reset).
	AppBaseURL string
	// CookieSecure marks session cookies Secure (set true behind HTTPS).
	CookieSecure bool

	// ResendAPIKey enables real email sending; empty falls back to console.
	ResendAPIKey string
	// MailFrom is the From address for outgoing emails.
	MailFrom string
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

	appBaseURL := os.Getenv("APP_BASE_URL")
	if appBaseURL == "" {
		appBaseURL = "http://localhost:5173"
	}

	mailFrom := os.Getenv("MAIL_FROM")
	if mailFrom == "" {
		mailFrom = "Nut Cracker <onboarding@resend.dev>"
	}

	return &Config{
		DatabaseURL:  dbURL,
		HTTPAddr:     httpAddr,
		AppBaseURL:   appBaseURL,
		CookieSecure: os.Getenv("COOKIE_SECURE") == "true",
		ResendAPIKey: os.Getenv("RESEND_API_KEY"),
		MailFrom:     mailFrom,
	}, nil
}
