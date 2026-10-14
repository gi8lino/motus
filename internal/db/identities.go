package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/gi8lino/motus/internal/utils"
)

// listAuthIdentities loads provider mappings keyed by their native Motus user.
func (s *Store) listAuthIdentities(ctx context.Context) (map[string][]AuthIdentity, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT user_id, provider, subject, email, display_name
		FROM auth_identities
		ORDER BY provider, subject
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	identities := make(map[string][]AuthIdentity)
	for rows.Next() {
		var userID string
		var identity AuthIdentity
		if err := rows.Scan(&userID, &identity.Provider, &identity.Subject, &identity.Email, &identity.Name); err != nil {
			return nil, err
		}
		identities[userID] = append(identities[userID], identity)
	}
	return identities, rows.Err()
}

// GetIdentityUser resolves a Motus user from a stable external identity.
func (s *Store) GetIdentityUser(ctx context.Context, provider, subject string) (*User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.is_admin, u.avatar_url, u.created_at
		FROM auth_identities i
		JOIN users u ON u.id = i.user_id
		WHERE i.provider = $1 AND i.subject = $2
	`, strings.TrimSpace(provider), strings.TrimSpace(subject))
	var user User
	if err := row.Scan(&user.ID, &user.Name, &user.IsAdmin, &user.AvatarURL, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

// ResolveExternalIdentity records a provider identity and returns its Motus user.
func (s *Store) ResolveExternalIdentity(
	ctx context.Context,
	provider, subject, email, name string,
	allowRegistration bool,
) (*User, error) {
	provider, subject = strings.TrimSpace(provider), strings.TrimSpace(subject)
	if provider == "" || subject == "" {
		return nil, errors.New("identity provider and subject are required")
	}
	if user, err := s.GetIdentityUser(ctx, provider, subject); err == nil {
		_, _ = s.pool.Exec(ctx, `
			UPDATE auth_identities
			SET email = $3, display_name = $4, last_seen_at = NOW()
			WHERE provider = $1 AND subject = $2
		`, provider, subject, strings.TrimSpace(email), strings.TrimSpace(name))
		return user, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if !allowRegistration {
		return nil, errors.New("user registration is disabled")
	}

	email, err := utils.NormalizeEmail(email)
	if err != nil {
		return nil, errors.New("identity provider did not supply a valid email")
	}
	// Never bind a newly seen external identity to an existing local account.
	// An administrator must link identities explicitly, so a mutable provider
	// email claim can never take over an existing account.
	if _, err := s.GetUser(ctx, email); err == nil {
		return nil, errors.New("an account already exists for this email")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	user := &User{ID: email, Name: strings.TrimSpace(name), CreatedAt: time.Now().UTC()}
	if user.Name == "" {
		user.Name = email
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO users(id, name, is_admin, avatar_url, password_hash, created_at)
		VALUES($1, $2, FALSE, '', '', $3)
	`, user.ID, user.Name, user.CreatedAt); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO auth_identities(provider, subject, user_id, email, display_name)
		VALUES($1, $2, $3, $4, $5)
	`, provider, subject, user.ID, email, user.Name); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetIdentityUser(ctx, provider, subject)
}
