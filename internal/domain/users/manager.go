package users

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/gi8lino/motus/internal/utils"
)

// Manager orchestrates user domain rules.
type Manager struct {
	store             Store
	authHeader        string
	allowRegistration bool
}

// NewManager creates a Manager instance.
func NewManager(store Store, authHeader string, allowRegistration bool) *Manager {
	return &Manager{store: store, authHeader: authHeader, allowRegistration: allowRegistration}
}

// Create enforces registration policy and persists a new user.
func (m *Manager) Create(ctx context.Context, email, avatarURL, password string) (*User, error) {
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

// List returns all users.
func (m *Manager) List(ctx context.Context) ([]User, error) {
	users, err := m.store.ListUsers(ctx)
	if err != nil {
		return nil, internal(err)
	}
	return users, nil
}

// Get returns a user by id.
func (m *Manager) Get(ctx context.Context, id string) (*User, error) {
	cleanID, err := requireEntityID(id, "user id is required")
	if err != nil {
		return nil, err
	}
	user, err := m.store.GetUser(ctx, cleanID)
	if err != nil {
		return nil, internal(err)
	}
	if user == nil {
		return nil, notFound("user not found")
	}
	return user, nil
}

// Login validates credentials when using local authentication.
func (m *Manager) Login(ctx context.Context, email, password string) (*User, error) {
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
