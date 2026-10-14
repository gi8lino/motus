package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// sessionTokenHash preserves legacy browser token hashes for existing sessions.
func sessionTokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// CreateSession stores a legacy opaque session token as its SHA-256 hash.
func (s *Store) CreateSession(ctx context.Context, token, userID string, expiresAt time.Time) error {
	return s.CreateSessionHash(ctx, sessionTokenHash(token), userID, expiresAt)
}

// CreateSessionHash stores a pre-hashed, purpose-separated session identifier.
func (s *Store) CreateSessionHash(ctx context.Context, hash, userID string, expiresAt time.Time) error {
	_, _ = s.pool.Exec(ctx, `DELETE FROM user_sessions WHERE expires_at <= NOW()`)
	_, err := s.pool.Exec(ctx, `INSERT INTO user_sessions(token, user_id, expires_at) VALUES ($1, $2, $3)`, hash, userID, expiresAt)
	return err
}

// GetSessionUser resolves a legacy session token.
func (s *Store) GetSessionUser(ctx context.Context, token string, now time.Time) (*User, error) {
	return s.GetSessionUserByHash(ctx, sessionTokenHash(token), now)
}

// GetSessionUserByHash resolves an unexpired purpose-separated token hash.
func (s *Store) GetSessionUserByHash(ctx context.Context, hash string, now time.Time) (*User, error) {
	row := s.pool.QueryRow(ctx, `
        SELECT u.id, u.name, u.is_admin, u.avatar_url, u.created_at
        FROM user_sessions s
        JOIN users u ON u.id = s.user_id
        WHERE s.token = $1 AND s.expires_at > $2
    `, hash, now)
	var user User
	if err := row.Scan(&user.ID, &user.Name, &user.IsAdmin, &user.AvatarURL, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteSession revokes a legacy opaque session token.
func (s *Store) DeleteSession(ctx context.Context, token string) error {
	return s.DeleteSessionHash(ctx, sessionTokenHash(token))
}

// DeleteSessionHash revokes a pre-hashed session token.
func (s *Store) DeleteSessionHash(ctx context.Context, hash string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM user_sessions WHERE token = $1`, hash)
	return err
}

// DeleteUserSessions revokes all browser sessions belonging to one user.
func (s *Store) DeleteUserSessions(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM user_sessions WHERE user_id = $1`, userID)
	return err
}
