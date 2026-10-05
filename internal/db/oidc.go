package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/gi8lino/motus/internal/utils"
)

// GetOIDCUser resolves a Motus user from a verified OpenID Connect identity.
func (s *Store) GetOIDCUser(ctx context.Context, issuer, subject string) (*User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.is_admin, u.avatar_url, u.created_at
		FROM oidc_identities i
		JOIN users u ON u.id = i.user_id
		WHERE i.issuer = $1 AND i.subject = $2
	`, strings.TrimSpace(issuer), strings.TrimSpace(subject))
	var user User
	if err := row.Scan(&user.ID, &user.Name, &user.IsAdmin, &user.AvatarURL, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}

// ResolveOIDCUser records a verified identity and returns its Motus user.
// New identities are admitted only when registration is enabled.
func (s *Store) ResolveOIDCUser(ctx context.Context, issuer, subject, email, name string, allowRegistration bool) (*User, error) {
	issuer, subject = strings.TrimSpace(issuer), strings.TrimSpace(subject)
	if issuer == "" || subject == "" {
		return nil, errors.New("oidc issuer and subject are required")
	}
	if user, err := s.GetOIDCUser(ctx, issuer, subject); err == nil {
		_, _ = s.pool.Exec(ctx, `UPDATE oidc_identities SET email=$3, display_name=$4, last_seen_at=NOW() WHERE issuer=$1 AND subject=$2`, issuer, subject, strings.TrimSpace(email), strings.TrimSpace(name))
		return user, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if !allowRegistration {
		return nil, errors.New("user registration is disabled")
	}

	email, err := utils.NormalizeEmail(email)
	if err != nil {
		return nil, errors.New("oidc provider did not supply a valid email")
	}
	// Never bind a newly seen external identity to an existing local account.
	// An administrator can choose a migration path explicitly; silently linking
	// by a mutable email claim would enable account takeover.
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
	if _, err := tx.Exec(ctx, `INSERT INTO users(id, name, is_admin, avatar_url, password_hash, created_at) VALUES($1,$2,FALSE,'','',$3)`, user.ID, user.Name, user.CreatedAt); err != nil {
		return nil, err
	}
	if err := tx.QueryRow(ctx, `SELECT id, name, is_admin, avatar_url, created_at FROM users WHERE id=$1`, user.ID).Scan(&user.ID, &user.Name, &user.IsAdmin, &user.AvatarURL, &user.CreatedAt); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO oidc_identities(issuer, subject, user_id, email, display_name) VALUES($1,$2,$3,$4,$5) ON CONFLICT (issuer, subject) DO UPDATE SET email=EXCLUDED.email, display_name=EXCLUDED.display_name, last_seen_at=NOW()`, issuer, subject, user.ID, email, user.Name); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetOIDCUser(ctx, issuer, subject)
}
