package db

import (
	"context"
	"time"
)

// CreateSession stores an opaque local-auth session token.
func (s *Store) CreateSession(ctx context.Context, token, userID string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO user_sessions(token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, token, userID, expiresAt)
	return err
}

// GetSessionUser resolves a live session and opportunistically ignores expired rows.
func (s *Store) GetSessionUser(ctx context.Context, token string, now time.Time) (*User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.is_admin, u.avatar_url, u.created_at
		FROM user_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = $1 AND s.expires_at > $2
	`, token, now)
	var user User
	if err := row.Scan(&user.ID, &user.Name, &user.IsAdmin, &user.AvatarURL, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}
