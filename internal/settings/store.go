// Package settings holds per-user feature flags / settings.
package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Settings holds a user's feature flags. Extend with one field per feature.
type Settings struct {
	CountMode bool `json:"count_mode"`
}

// Store provides access to user settings.
type Store struct {
	db *pgxpool.Pool
}

// NewStore creates a Store backed by the given pool.
func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// Get returns the user's settings, or defaults if none have been saved yet.
func (s *Store) Get(ctx context.Context, userID int64) (Settings, error) {
	var out Settings
	err := s.db.QueryRow(ctx,
		`SELECT count_mode FROM user_settings WHERE user_id = $1`, userID).
		Scan(&out.CountMode)
	if errors.Is(err, pgx.ErrNoRows) {
		return Settings{}, nil // defaults (all features off)
	}
	if err != nil {
		return Settings{}, fmt.Errorf("querying settings: %w", err)
	}
	return out, nil
}

// Update upserts the user's settings and returns the stored values.
func (s *Store) Update(ctx context.Context, userID int64, in Settings) (Settings, error) {
	var out Settings
	err := s.db.QueryRow(ctx,
		`INSERT INTO user_settings (user_id, count_mode)
		 VALUES ($1, $2)
		 ON CONFLICT (user_id)
		 DO UPDATE SET count_mode = EXCLUDED.count_mode, updated_at = now()
		 RETURNING count_mode`,
		userID, in.CountMode).Scan(&out.CountMode)
	if err != nil {
		return Settings{}, fmt.Errorf("updating settings: %w", err)
	}
	return out, nil
}
