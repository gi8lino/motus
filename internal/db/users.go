package db

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CreateUser inserts a new user with the provided password hash.
func (s *Store) CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*User, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	user := &User{
		ID:        normalized,
		Name:      normalized,
		IsAdmin:   false,
		AvatarURL: strings.TrimSpace(avatarURL),
		CreatedAt: time.Now().UTC(),
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO users(
			id,
			name,
			is_admin,
			avatar_url,
			password_hash,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		user.ID, user.Name, user.IsAdmin, user.AvatarURL, strings.TrimSpace(passwordHash), user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ListUsers returns all users ordered by creation date.
func (s *Store) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, is_admin, avatar_url, created_at
		FROM users
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.IsAdmin, &u.AvatarURL, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// GetUser fetches a single user by id.
func (s *Store) GetUser(ctx context.Context, id string) (*User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, is_admin, avatar_url, created_at
		FROM users
		WHERE id=$1
	`, strings.TrimSpace(id))
	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.IsAdmin, &u.AvatarURL, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserWithPassword fetches a user and password hash by id.
func (s *Store) GetUserWithPassword(ctx context.Context, id string) (*User, string, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, name, is_admin, avatar_url, created_at, password_hash
		FROM users
		WHERE id=$1
	`, strings.TrimSpace(id))
	var u User
	var passwordHash string
	if err := row.Scan(&u.ID, &u.Name, &u.IsAdmin, &u.AvatarURL, &u.CreatedAt, &passwordHash); err != nil {
		return nil, "", err
	}
	return &u, passwordHash, nil
}

// UpdateUserPassword sets the password hash for a user.
func (s *Store) UpdateUserPassword(ctx context.Context, userID, passwordHash string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE users
		SET password_hash=$1
		WHERE id=$2
	`, strings.TrimSpace(passwordHash), strings.TrimSpace(userID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// UpdateUserAdmin sets the admin flag for a user.
func (s *Store) UpdateUserAdmin(ctx context.Context, userID string, isAdmin bool) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE users
		SET is_admin=$1
		WHERE id=$2
	`, isAdmin, strings.TrimSpace(userID))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// UpsertAdminUser ensures the admin user exists with the given password hash.
func (s *Store) UpsertAdminUser(ctx context.Context, email, passwordHash string) (*User, bool, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return nil, false, fmt.Errorf("admin email is required")
	}
	if strings.TrimSpace(passwordHash) == "" {
		return nil, false, fmt.Errorf("admin password hash is required")
	}
	now := time.Now().UTC()
	row := s.pool.QueryRow(
		ctx,
		`
			INSERT INTO users(
				id,
				name,
				is_admin,
				avatar_url,
				password_hash,
				created_at
			)
			VALUES ($1, $2, TRUE, '', $3, $4)
			ON CONFLICT (id) DO UPDATE
			SET is_admin=TRUE,
				password_hash=EXCLUDED.password_hash
			RETURNING id, name, is_admin, avatar_url, created_at, (xmax = 0) AS created
		`,
		normalized,
		normalized,
		strings.TrimSpace(passwordHash),
		now,
	)
	var u User
	var created bool
	if err := row.Scan(&u.ID, &u.Name, &u.IsAdmin, &u.AvatarURL, &u.CreatedAt, &created); err != nil {
		return nil, false, err
	}
	return &u, created, nil
}
