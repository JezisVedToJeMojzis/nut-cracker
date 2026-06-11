// Package server wires HTTP routing and handlers for the Nut Cracker API.
package server

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Server holds dependencies shared across HTTP handlers.
type Server struct {
	db *pgxpool.Pool
}

// New creates a Server with the given database pool.
func New(db *pgxpool.Pool) *Server {
	return &Server{db: db}
}

// Routes returns the HTTP handler with all routes registered.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	return mux
}

// handleHealth reports service and database health.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	code := http.StatusOK

	if err := s.db.Ping(r.Context()); err != nil {
		status = "database unreachable"
		code = http.StatusServiceUnavailable
	}

	writeJSON(w, code, map[string]string{"status": status})
}

// writeJSON writes v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
