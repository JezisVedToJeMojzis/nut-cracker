// Package friends holds domain logic for friend requests and relationships.
package friends

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrSelf is returned when a user targets themselves.
	ErrSelf = errors.New("cannot befriend yourself")
	// ErrUserNotFound is returned when the target user does not exist.
	ErrUserNotFound = errors.New("user not found")
	// ErrAlreadyFriends is returned when an accepted friendship already exists.
	ErrAlreadyFriends = errors.New("already friends")
	// ErrRequestExists is returned when a pending request already exists.
	ErrRequestExists = errors.New("friend request already pending")
	// ErrRequestNotFound is returned when no matching pending request exists.
	ErrRequestNotFound = errors.New("no pending request found")
	// ErrNotFound is returned when no relationship exists to remove.
	ErrNotFound = errors.New("relationship not found")
)

// Store provides access to friendship data.
type Store struct {
	db *pgxpool.Pool
}

// NewStore creates a Store backed by the given pool.
func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// User is a minimal user reference returned in listings.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// SendRequest creates a pending request from -> to. If a reciprocal pending
// request (to -> from) already exists, it is accepted instead and accepted=true
// is returned.
func (s *Store) SendRequest(ctx context.Context, from, to string) (accepted bool, err error) {
	if from == to {
		return false, ErrSelf
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Target must exist.
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`, to).Scan(&exists); err != nil {
		return false, fmt.Errorf("checking user: %w", err)
	}
	if !exists {
		return false, ErrUserNotFound
	}

	// Inspect any existing relationship in either direction.
	var dir string // 'from_to' or 'to_from'
	var status string
	err = tx.QueryRow(ctx,
		`SELECT CASE WHEN user_id = $1 THEN 'from_to' ELSE 'to_from' END, status
		 FROM friendships
		 WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)`,
		from, to).Scan(&dir, &status)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		// No relationship yet: create the pending request.
		if _, err := tx.Exec(ctx,
			`INSERT INTO friendships (user_id, friend_id, status) VALUES ($1, $2, 'pending')`,
			from, to); err != nil {
			return false, fmt.Errorf("inserting request: %w", err)
		}
	case err != nil:
		return false, fmt.Errorf("checking relationship: %w", err)
	case status == "accepted":
		return false, ErrAlreadyFriends
	case dir == "from_to":
		// We already sent a pending request.
		return false, ErrRequestExists
	default:
		// dir == "to_from": they already requested us -> accept it.
		if _, err := tx.Exec(ctx,
			`UPDATE friendships SET status = 'accepted'
			 WHERE user_id = $1 AND friend_id = $2`,
			to, from); err != nil {
			return false, fmt.Errorf("accepting reciprocal: %w", err)
		}
		accepted = true
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit: %w", err)
	}
	return accepted, nil
}

// Accept marks a pending request (requester -> recipient) as accepted.
func (s *Store) Accept(ctx context.Context, recipient, requester string) error {
	tag, err := s.db.Exec(ctx,
		`UPDATE friendships SET status = 'accepted'
		 WHERE user_id = $1 AND friend_id = $2 AND status = 'pending'`,
		requester, recipient)
	if err != nil {
		return fmt.Errorf("accepting: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrRequestNotFound
	}
	return nil
}

// Remove deletes any relationship between the two users (accepted or pending),
// covering unfriend, cancel-request, and decline-request.
func (s *Store) Remove(ctx context.Context, userID, otherID string) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM friendships
		 WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)`,
		userID, otherID)
	if err != nil {
		return fmt.Errorf("removing: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListFriends returns accepted friends of the user (either direction).
func (s *Store) ListFriends(ctx context.Context, userID string) ([]User, error) {
	return s.queryUsers(ctx,
		`SELECT u.id, u.username FROM friendships f
		 JOIN users u ON u.id = CASE WHEN f.user_id = $1 THEN f.friend_id ELSE f.user_id END
		 WHERE f.status = 'accepted' AND (f.user_id = $1 OR f.friend_id = $1)
		 ORDER BY u.username`, userID)
}

// ListIncoming returns users who have sent the user a pending request.
func (s *Store) ListIncoming(ctx context.Context, userID string) ([]User, error) {
	return s.queryUsers(ctx,
		`SELECT u.id, u.username FROM friendships f
		 JOIN users u ON u.id = f.user_id
		 WHERE f.status = 'pending' AND f.friend_id = $1
		 ORDER BY u.username`, userID)
}

// ListOutgoing returns users the user has sent a pending request to.
func (s *Store) ListOutgoing(ctx context.Context, userID string) ([]User, error) {
	return s.queryUsers(ctx,
		`SELECT u.id, u.username FROM friendships f
		 JOIN users u ON u.id = f.friend_id
		 WHERE f.status = 'pending' AND f.user_id = $1
		 ORDER BY u.username`, userID)
}

func (s *Store) queryUsers(ctx context.Context, sql, userID string) ([]User, error) {
	rows, err := s.db.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}
