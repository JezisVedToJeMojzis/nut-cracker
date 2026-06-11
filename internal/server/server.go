// Package server wires HTTP routing and handlers for the Nut Cracker API.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"nutcracker/internal/friends"
	"nutcracker/internal/settings"
	"nutcracker/internal/usermap"
	"nutcracker/internal/users"
)

// Server holds dependencies shared across HTTP handlers.
type Server struct {
	db       *pgxpool.Pool
	maps     *usermap.Store
	friends  *friends.Store
	settings *settings.Store
	users    *users.Store
}

// New creates a Server with the given database pool.
func New(db *pgxpool.Pool) *Server {
	return &Server{
		db:       db,
		maps:     usermap.NewStore(db),
		friends:  friends.NewStore(db),
		settings: settings.NewStore(db),
		users:    users.NewStore(db),
	}
}

// Routes returns the HTTP handler with all routes registered.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)

	mux.HandleFunc("GET /users/{id}", s.handleGetProfile)
	mux.HandleFunc("PATCH /users/{id}", s.handleUpdateProfile)
	mux.HandleFunc("GET /users/{id}/card", s.handleGetCard)

	mux.HandleFunc("GET /users/{id}/map", s.handleGetMap)
	mux.HandleFunc("POST /users/{id}/countries/{code}/increment", s.handleIncrement)
	mux.HandleFunc("POST /users/{id}/countries/{code}/decrement", s.handleDecrement)
	mux.HandleFunc("DELETE /users/{id}/countries/{code}", s.handleRemove)

	mux.HandleFunc("POST /friends/requests", s.handleSendRequest)
	mux.HandleFunc("POST /friends/requests/{requesterId}/accept", s.handleAcceptRequest)
	mux.HandleFunc("POST /friends/requests/{requesterId}/decline", s.handleDeclineRequest)
	mux.HandleFunc("GET /friends", s.handleListFriends)
	mux.HandleFunc("GET /friends/requests/incoming", s.handleListIncoming)
	mux.HandleFunc("GET /friends/requests/outgoing", s.handleListOutgoing)
	mux.HandleFunc("DELETE /friends/{otherId}", s.handleRemoveFriend)

	mux.HandleFunc("GET /users/{id}/settings", s.handleGetSettings)
	mux.HandleFunc("PUT /users/{id}/settings", s.handleUpdateSettings)
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

// handleGetProfile returns the caller's full profile. Self only.
func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireUser(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if caller != id {
		writeError(w, http.StatusForbidden, "cannot view another user's profile")
		return
	}
	p, err := s.users.GetByID(r.Context(), id)
	if errors.Is(err, users.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// handleUpdateProfile updates the caller's username. Self only.
func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireUser(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if caller != id {
		writeError(w, http.StatusForbidden, "cannot edit another user's profile")
		return
	}
	var body struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	p, err := s.users.UpdateUsername(r.Context(), id, body.Username)
	switch {
	case errors.Is(err, users.ErrInvalidUsername):
		writeError(w, http.StatusBadRequest, "username must be 2-30 characters")
	case errors.Is(err, users.ErrUsernameTaken):
		writeError(w, http.StatusConflict, "username already taken")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusOK, p)
	}
}

// handleGetCard returns a user's public card (id + username) for search.
func (s *Server) handleGetCard(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireUser(w, r); !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	u, err := s.users.LookupByID(r.Context(), id)
	if errors.Is(err, users.ErrNotFound) {
		writeError(w, http.StatusNotFound, "no user with that id")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// handleGetMap returns a user's map, enforcing friends-only visibility.
func (s *Server) handleGetMap(w http.ResponseWriter, r *http.Request) {
	viewer, ok := requireUser(w, r)
	if !ok {
		return
	}
	ownerID, ok := pathID(w, r, "id")
	if !ok {
		return
	}

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

// handleSendRequest sends a friend request from the caller to body.To.
func (s *Server) handleSendRequest(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireUser(w, r)
	if !ok {
		return
	}

	var body struct {
		To int64 `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.To == 0 {
		writeError(w, http.StatusBadRequest, "expected JSON body with a 'to' user id")
		return
	}

	accepted, err := s.friends.SendRequest(r.Context(), caller, body.To)
	switch {
	case errors.Is(err, friends.ErrSelf):
		writeError(w, http.StatusBadRequest, "cannot befriend yourself")
	case errors.Is(err, friends.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, friends.ErrAlreadyFriends):
		writeError(w, http.StatusConflict, "already friends")
	case errors.Is(err, friends.ErrRequestExists):
		writeError(w, http.StatusConflict, "request already pending")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		status := "pending"
		if accepted {
			status = "accepted"
		}
		writeJSON(w, http.StatusCreated, map[string]string{"status": status})
	}
}

// handleAcceptRequest accepts a pending request addressed to the caller.
func (s *Server) handleAcceptRequest(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireUser(w, r)
	if !ok {
		return
	}
	requester, ok := pathID(w, r, "requesterId")
	if !ok {
		return
	}

	err := s.friends.Accept(r.Context(), caller, requester)
	switch {
	case errors.Is(err, friends.ErrRequestNotFound):
		writeError(w, http.StatusNotFound, "no pending request from that user")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
	}
}

// handleDeclineRequest declines a pending request addressed to the caller.
func (s *Server) handleDeclineRequest(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireUser(w, r)
	if !ok {
		return
	}
	requester, ok := pathID(w, r, "requesterId")
	if !ok {
		return
	}

	err := s.friends.Decline(r.Context(), caller, requester)
	switch {
	case errors.Is(err, friends.ErrRequestNotFound):
		writeError(w, http.StatusNotFound, "no pending request from that user")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusOK, map[string]string{"status": "declined"})
	}
}

// handleRemoveFriend removes any relationship between caller and {otherId}.
func (s *Server) handleRemoveFriend(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireUser(w, r)
	if !ok {
		return
	}
	other, ok := pathID(w, r, "otherId")
	if !ok {
		return
	}
	err := s.friends.Remove(r.Context(), caller, other)
	switch {
	case errors.Is(err, friends.ErrNotFound):
		writeError(w, http.StatusNotFound, "no relationship to remove")
	case err != nil:
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) handleListFriends(w http.ResponseWriter, r *http.Request) {
	s.listFriendsBy(w, r, s.friends.ListFriends, "friends")
}

func (s *Server) handleListIncoming(w http.ResponseWriter, r *http.Request) {
	s.listFriendsBy(w, r, s.friends.ListIncoming, "requests")
}

func (s *Server) handleListOutgoing(w http.ResponseWriter, r *http.Request) {
	s.listFriendsBy(w, r, s.friends.ListOutgoing, "requests")
}

// listFriendsBy runs a listing function for the caller and writes the result
// under the given JSON key.
func (s *Server) listFriendsBy(w http.ResponseWriter, r *http.Request, list func(context.Context, int64) ([]friends.User, error), key string) {
	caller, ok := requireUser(w, r)
	if !ok {
		return
	}
	list2, err := list(r.Context(), caller)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if list2 == nil {
		list2 = []friends.User{}
	}
	writeJSON(w, http.StatusOK, map[string]any{key: list2})
}

// handleGetSettings returns the caller's settings. Self only.
func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireUser(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if caller != id {
		writeError(w, http.StatusForbidden, "cannot read another user's settings")
		return
	}
	st, err := s.settings.Get(r.Context(), caller)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// handleUpdateSettings upserts the caller's settings. Self only.
func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireUser(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if caller != id {
		writeError(w, http.StatusForbidden, "cannot modify another user's settings")
		return
	}
	var body settings.Settings
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	st, err := s.settings.Update(r.Context(), caller, body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// requireSelf checks that the caller is acting on their own map. It returns the
// user ID, country code, and ok=false (after writing an error) when the caller
// is unauthenticated or targeting someone else's map.
func (s *Server) requireSelf(w http.ResponseWriter, r *http.Request) (userID int64, code string, ok bool) {
	viewer, ok := requireUser(w, r)
	if !ok {
		return 0, "", false
	}
	ownerID, ok := pathID(w, r, "id")
	if !ok {
		return 0, "", false
	}
	if viewer != ownerID {
		writeError(w, http.StatusForbidden, "cannot modify another user's map")
		return 0, "", false
	}
	return ownerID, r.PathValue("code"), true
}

// requireUser extracts and validates the acting user's id from the X-User-ID
// header. Temporary stand-in for real auth (Google OAuth / sessions).
func requireUser(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := r.Header.Get("X-User-ID")
	if raw == "" {
		writeError(w, http.StatusUnauthorized, "missing X-User-ID")
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid X-User-ID")
		return 0, false
	}
	return id, true
}

// pathID parses a numeric path value, writing a 400 on failure.
func pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid "+name)
		return 0, false
	}
	return id, true
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
