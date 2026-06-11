// Package usermap holds domain logic for a user's map of cracked countries
// and the friendship-based visibility rules around it.
package usermap

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrCountryNotFound is returned when an operation targets a country the user
// has not cracked.
var ErrCountryNotFound = errors.New("country not on user's map")

// Store provides access to user map data.
type Store struct {
	db *pgxpool.Pool
}

// NewStore creates a Store backed by the given pool.
func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// CountryCount is one entry on a user's map.
type CountryCount struct {
	CountryCode string `json:"country_code"`
	Cracks      int    `json:"cracks"`
}

// List returns all countries the given user has cracked.
func (s *Store) List(ctx context.Context, userID string) ([]CountryCount, error) {
	rows, err := s.db.Query(ctx,
		`SELECT country_code, cracks FROM user_countries
		 WHERE user_id = $1 ORDER BY country_code`, userID)
	if err != nil {
		return nil, fmt.Errorf("querying map: %w", err)
	}
	defer rows.Close()

	var out []CountryCount
	for rows.Next() {
		var c CountryCount
		if err := rows.Scan(&c.CountryCode, &c.Cracks); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Increment adds one crack for the country, creating the entry at 1 if absent.
// Returns the new crack count.
func (s *Store) Increment(ctx context.Context, userID, code string) (int, error) {
	var cracks int
	err := s.db.QueryRow(ctx,
		`INSERT INTO user_countries (user_id, country_code, cracks)
		 VALUES ($1, $2, 1)
		 ON CONFLICT (user_id, country_code)
		 DO UPDATE SET cracks = user_countries.cracks + 1, updated_at = now()
		 RETURNING cracks`,
		userID, code).Scan(&cracks)
	if err != nil {
		return 0, fmt.Errorf("incrementing: %w", err)
	}
	return cracks, nil
}

// Decrement removes one crack. When the count would drop to zero the country is
// removed from the map. Returns the new count and whether the country was
// removed. Returns ErrCountryNotFound if the user has not cracked the country.
func (s *Store) Decrement(ctx context.Context, userID, code string) (newCount int, removed bool, err error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var current int
	err = tx.QueryRow(ctx,
		`SELECT cracks FROM user_countries
		 WHERE user_id = $1 AND country_code = $2 FOR UPDATE`,
		userID, code).Scan(&current)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, ErrCountryNotFound
	}
	if err != nil {
		return 0, false, fmt.Errorf("selecting current: %w", err)
	}

	if current <= 1 {
		if _, err := tx.Exec(ctx,
			`DELETE FROM user_countries WHERE user_id = $1 AND country_code = $2`,
			userID, code); err != nil {
			return 0, false, fmt.Errorf("deleting: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return 0, false, fmt.Errorf("commit: %w", err)
		}
		return 0, true, nil
	}

	if err := tx.QueryRow(ctx,
		`UPDATE user_countries SET cracks = cracks - 1, updated_at = now()
		 WHERE user_id = $1 AND country_code = $2 RETURNING cracks`,
		userID, code).Scan(&newCount); err != nil {
		return 0, false, fmt.Errorf("updating: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, false, fmt.Errorf("commit: %w", err)
	}
	return newCount, false, nil
}

// Remove deletes a country from the user's map entirely. Returns
// ErrCountryNotFound if it was not present.
func (s *Store) Remove(ctx context.Context, userID, code string) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM user_countries WHERE user_id = $1 AND country_code = $2`,
		userID, code)
	if err != nil {
		return fmt.Errorf("deleting: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCountryNotFound
	}
	return nil
}

// CanView reports whether viewer is allowed to see owner's map: true if they are
// the same user or are accepted friends in either direction.
func (s *Store) CanView(ctx context.Context, viewerID, ownerID string) (bool, error) {
	if viewerID == ownerID {
		return true, nil
	}
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS (
		   SELECT 1 FROM friendships
		   WHERE status = 'accepted'
		     AND ((user_id = $1 AND friend_id = $2)
		       OR (user_id = $2 AND friend_id = $1))
		 )`, viewerID, ownerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking friendship: %w", err)
	}
	return exists, nil
}
