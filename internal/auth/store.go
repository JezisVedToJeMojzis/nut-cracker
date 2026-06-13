// Package auth handles credentials, DB-backed sessions, and one-time tokens
// for email verification and password reset.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrEmailTaken is returned when registering with an existing email.
	ErrEmailTaken = errors.New("email already registered")
	// ErrInvalidCredentials is returned on a failed login.
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrInvalidInput is returned when registration input fails validation.
	ErrInvalidInput = errors.New("invalid input")
	// ErrNoSession is returned when a session token is missing/expired.
	ErrNoSession = errors.New("no valid session")
	// ErrInvalidToken is returned for a missing/expired one-time token.
	ErrInvalidToken = errors.New("invalid or expired token")
)

// Store provides authentication data access.
type Store struct {
	db *pgxpool.Pool
}

// NewStore creates a Store backed by the given pool.
func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Register creates a new account and a verification token. The password is
// bcrypt-hashed. Returns the new user id and the verification token.
func (s *Store) Register(ctx context.Context, email, username, password string) (userID int64, verifyToken string, err error) {
	email = strings.ToLower(strings.TrimSpace(email))
	username = strings.TrimSpace(username)
	if !strings.Contains(email, "@") || len(username) < 2 || len(username) > 30 || len(password) < 8 {
		return 0, "", ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, "", fmt.Errorf("hashing password: %w", err)
	}

	err = s.db.QueryRow(ctx,
		`INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		email, username, string(hash)).Scan(&userID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return 0, "", ErrEmailTaken
	}
	if err != nil {
		return 0, "", fmt.Errorf("inserting user: %w", err)
	}

	verifyToken, err = s.CreateToken(ctx, userID, "verify", 48*time.Hour)
	if err != nil {
		return 0, "", err
	}
	return userID, verifyToken, nil
}

// Authenticate verifies an email/password pair and returns the user id.
func (s *Store) Authenticate(ctx context.Context, email, password string) (int64, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var id int64
	var hash *string
	err := s.db.QueryRow(ctx,
		`SELECT id, password_hash FROM users WHERE email = $1`, email).Scan(&id, &hash)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && hash == nil) {
		return 0, ErrInvalidCredentials
	}
	if err != nil {
		return 0, fmt.Errorf("querying user: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(*hash), []byte(password)) != nil {
		return 0, ErrInvalidCredentials
	}
	return id, nil
}

// CreateSession creates a session valid for the given duration.
func (s *Store) CreateSession(ctx context.Context, userID int64, ttl time.Duration) (token string, expires time.Time, err error) {
	token, err = randomToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expires = time.Now().Add(ttl)
	_, err = s.db.Exec(ctx,
		`INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)`,
		token, userID, expires)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("creating session: %w", err)
	}
	return token, expires, nil
}

// UserIDForSession returns the user id for a valid (unexpired) session token.
func (s *Store) UserIDForSession(ctx context.Context, token string) (int64, error) {
	if token == "" {
		return 0, ErrNoSession
	}
	var id int64
	err := s.db.QueryRow(ctx,
		`SELECT user_id FROM sessions WHERE id = $1 AND expires_at > now()`, token).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNoSession
	}
	if err != nil {
		return 0, fmt.Errorf("querying session: %w", err)
	}
	return id, nil
}

// DeleteSession removes a session (logout).
func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, token)
	return err
}

// CreateToken creates a one-time token of the given kind.
func (s *Store) CreateToken(ctx context.Context, userID int64, kind string, ttl time.Duration) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	_, err = s.db.Exec(ctx,
		`INSERT INTO user_tokens (token, user_id, kind, expires_at) VALUES ($1, $2, $3, $4)`,
		token, userID, kind, time.Now().Add(ttl))
	if err != nil {
		return "", fmt.Errorf("creating token: %w", err)
	}
	return token, nil
}

// ConsumeToken validates and deletes a one-time token, returning its user id.
func (s *Store) ConsumeToken(ctx context.Context, token, kind string) (int64, error) {
	var userID int64
	err := s.db.QueryRow(ctx,
		`DELETE FROM user_tokens
		 WHERE token = $1 AND kind = $2 AND expires_at > now()
		 RETURNING user_id`,
		token, kind).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrInvalidToken
	}
	if err != nil {
		return 0, fmt.Errorf("consuming token: %w", err)
	}
	return userID, nil
}

// MarkEmailVerified flags a user's email as verified.
func (s *Store) MarkEmailVerified(ctx context.Context, userID int64) error {
	_, err := s.db.Exec(ctx, `UPDATE users SET email_verified = true WHERE id = $1`, userID)
	return err
}

// SetPassword updates a user's password (used by reset).
func (s *Store) SetPassword(ctx context.Context, userID int64, password string) error {
	if len(password) < 8 {
		return ErrInvalidInput
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}
	_, err = s.db.Exec(ctx, `UPDATE users SET password_hash = $2 WHERE id = $1`, userID, string(hash))
	return err
}

// UserIDByEmail returns the user id for an email, or ok=false if none.
func (s *Store) UserIDByEmail(ctx context.Context, email string) (int64, bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var id int64
	err := s.db.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("querying email: %w", err)
	}
	return id, true, nil
}
