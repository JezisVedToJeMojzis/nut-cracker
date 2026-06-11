// Package users holds domain logic for user profiles and lookup/search.
package users

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrNotFound is returned when no user matches.
	ErrNotFound = errors.New("user not found")
	// ErrInvalidUsername is returned when a username fails validation.
	ErrInvalidUsername = errors.New("username must be 2-30 characters")
)

// Store provides access to user data.
type Store struct {
	db *pgxpool.Pool
}

// NewStore creates a Store backed by the given pool.
func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// Profile is a user's own full account info (includes private fields).
type Profile struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// PublicUser is the minimal, public-facing view used for search results.
type PublicUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

// GetByID returns a user's full profile.
func (s *Store) GetByID(ctx context.Context, id int64) (Profile, error) {
	var p Profile
	err := s.db.QueryRow(ctx,
		`SELECT id, username, email, created_at FROM users WHERE id = $1`, id).
		Scan(&p.ID, &p.Username, &p.Email, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, ErrNotFound
	}
	if err != nil {
		return Profile{}, fmt.Errorf("querying profile: %w", err)
	}
	return p, nil
}

// LookupByID returns the public view of a user by their numeric id (used by
// friend search so the caller can see the username before adding).
func (s *Store) LookupByID(ctx context.Context, id int64) (PublicUser, error) {
	var u PublicUser
	err := s.db.QueryRow(ctx,
		`SELECT id, username FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Username)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicUser{}, ErrNotFound
	}
	if err != nil {
		return PublicUser{}, fmt.Errorf("looking up user: %w", err)
	}
	return u, nil
}

// UpdateUsername changes a user's display name and returns the updated profile.
func (s *Store) UpdateUsername(ctx context.Context, id int64, username string) (Profile, error) {
	username = strings.TrimSpace(username)
	if len(username) < 2 || len(username) > 30 {
		return Profile{}, ErrInvalidUsername
	}

	var p Profile
	err := s.db.QueryRow(ctx,
		`UPDATE users SET username = $2 WHERE id = $1
		 RETURNING id, username, email, created_at`,
		id, username).Scan(&p.ID, &p.Username, &p.Email, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, ErrNotFound
	}
	if err != nil {
		return Profile{}, fmt.Errorf("updating username: %w", err)
	}
	return p, nil
}
