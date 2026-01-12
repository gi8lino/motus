package users

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

// Login validates credentials when using local authentication.
func (m *Manager) Login(ctx context.Context, email, password string) (*db.User, error) {
	if m.authHeader != "" {
		return nil, forbidden("local login disabled")
	}

	normalized, err := utils.NormalizeEmail(email)
	if err != nil {
		return nil, validation(err.Error())
	}

	password = strings.TrimSpace(password)
	if password == "" {
		return nil, validation("email and password are required")
	}

	user, hash, err := m.store.GetUserWithPassword(ctx, normalized)
	if err != nil {
		return nil, internal(err)
	}
	if user == nil || hash == "" {
		return nil, unauthorized("invalid credentials")
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, unauthorized("invalid credentials")
	}

	return user, nil
}

// ChangePassword updates the password for the specified user.
func (m *Manager) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if m.authHeader != "" {
		return forbidden("passwords managed by proxy")
	}

	currentPassword = strings.TrimSpace(currentPassword)
	newPassword = strings.TrimSpace(newPassword)
	if currentPassword == "" || newPassword == "" {
		return validation("current and new password are required")
	}

	user, hash, err := m.store.GetUserWithPassword(ctx, userID)
	if err != nil {
		return internal(err)
	}
	if user == nil || hash == "" {
		return unauthorized("invalid credentials")
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword)) != nil {
		return unauthorized("invalid credentials")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return internal(fmt.Errorf("unable to secure password: %w", err))
	}

	if err := m.store.UpdateUserPassword(ctx, userID, string(newHash)); err != nil {
		return internal(err)
	}
	return nil
}
