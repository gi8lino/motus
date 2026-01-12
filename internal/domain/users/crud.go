package users

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

// Create enforces registration policy and persists a new user.
func (m *Manager) Create(ctx context.Context, email, avatarURL, password string) (*db.User, error) {
	normalized, err := utils.NormalizeEmail(email)
	if err != nil {
		return nil, validation(err.Error())
	}

	if m.authHeader == "" && !m.allowRegistration {
		return nil, forbidden("registration is disabled")
	}

	password = strings.TrimSpace(password)
	if m.authHeader == "" && password == "" {
		return nil, validation("password is required")
	}

	var passwordHash string
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, internal(fmt.Errorf("unable to secure password: %w", err))
		}
		passwordHash = string(hash)
	}

	user, err := m.store.CreateUser(ctx, normalized, avatarURL, passwordHash)
	if err != nil {
		return nil, internal(err)
	}
	return user, nil
}

// UpdateRole toggles the admin flag.
func (m *Manager) UpdateRole(ctx context.Context, id string, isAdmin bool) error {
	cleanID, err := requireEntityID(id, "user id is required")
	if err != nil {
		return err
	}
	if err := m.store.UpdateUserAdmin(ctx, cleanID, isAdmin); err != nil {
		return internal(err)
	}
	return nil
}

// UpdateName changes the user's display name.
func (m *Manager) UpdateName(ctx context.Context, userID, name string) error {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return validation("name is required")
	}
	if err := m.store.UpdateUserName(ctx, userID, cleanName); err != nil {
		return internal(err)
	}
	return nil
}
