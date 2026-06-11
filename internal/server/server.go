// Package server wires HTTP routing and handlers for the Nut Cracker API.
package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"nutcracker/internal/usermap"
)

// Server holds dependencies shared across HTTP handlers.
type Server struct {
	db   *pgxpool.Pool
	maps *usermap.Store
}

// New creates a Server with the given database pool.
func New(db *pgxpool.Pool) *Server {
	return &Server{
		db:   db,
		maps: usermap.NewStore(db),
	}
}

// Routes returns the HTTP handler with all routes registered.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /users/{id}/map", s.handleGetMap)
	mux.HandleFunc("POST /users/{id}/countries/{code}/increment", s.handleIncrement)
	mux.HandleFunc("POST /users/{id}/countries/{code}/decrement", s.handleDecrement)
	mux.HandleFunc("DELETE /users/{id}/countries/{code}", s.handleRemove)
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

// handleGetMap returns a user's map, enforcing friends-only visibility.
func (s *Server) handleGetMap(w http.ResponseWriter, r *http.Request) {
	viewer := currentUser(r)
	if viewer == "" {
		writeError(w, http.StatusUnauthorized, "missing X-User-ID")
		return
	}
	ownerID := r.PathValue("id")

	allowed, err := s.maps.CanView(r.Context(), viewer, ownerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !allowed {
		writeError(w, http.StatusForbidden, "not allowed to view this map")
		return
	}

	countries, err := s.maps.List(r.Context(), ownerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if countries == nil {
		countries = []usermap.CountryCount{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"countries": countries})
}

// handleIncrement adds one crack to a country on the caller's own map.
func (s *Server) handleIncrement(w http.ResponseWriter, r *http.Request) {
	userID, code, ok := s.requireSelf(w, r)
	if !ok {
		return
	}
	cracks, err := s.maps.Increment(r.Context(), userID, code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"country_code": code, "cracks": cracks})
}

// handleDecrement removes one crack; removes the country when it reaches zero.
func (s *Server) handleDecrement(w http.ResponseWriter, r *http.Request) {
	userID, code, ok := s.requireSelf(w, r)
	if !ok {
		return
	}
	cracks, removed, err := s.maps.Decrement(r.Context(), userID, code)
	if errors.Is(err, usermap.ErrCountryNotFound) {
		writeError(w, http.StatusNotFound, "country not on your map")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"country_code": code,
		"cracks":       cracks,
		"removed":      removed,
	})
}

// handleRemove uncolors a country entirely from the caller's own map.
func (s *Server) handleRemove(w http.ResponseWriter, r *http.Request) {
	userID, code, ok := s.requireSelf(w, r)
	if !ok {
		return
	}
	err := s.maps.Remove(r.Context(), userID, code)
	if errors.Is(err, usermap.ErrCountryNotFound) {
		writeError(w, http.StatusNotFound, "country not on your map")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// requireSelf checks that the caller is acting on their own map. It returns the
// user ID, country code, and ok=false (after writing an error) when the caller
// is unauthenticated or targeting someone else's map.
func (s *Server) requireSelf(w http.ResponseWriter, r *http.Request) (userID, code string, ok bool) {
	viewer := currentUser(r)
	if viewer == "" {
		writeError(w, http.StatusUnauthorized, "missing X-User-ID")
		return "", "", false
	}
	ownerID := r.PathValue("id")
	if viewer != ownerID {
		writeError(w, http.StatusForbidden, "cannot modify another user's map")
		return "", "", false
	}
	return ownerID, r.PathValue("code"), true
}

// currentUser extracts the acting user's ID. Temporary stand-in for real auth:
// reads the X-User-ID header. To be replaced by Google OAuth / sessions.
func currentUser(r *http.Request) string {
	return r.Header.Get("X-User-ID")
}

// writeJSON writes v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error body with the given status code.
func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
