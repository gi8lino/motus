package users

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
	"github.com/gi8lino/motus/internal/utils"
)

// Create registers a new local user.
func (s *Service) Create(ctx context.Context, email, avatarURL, password string) (*User, error) {
	return s.create(ctx, email, avatarURL, password, true)
}

// CreateManaged creates a local user through administrator-managed user administration.
func (s *Service) CreateManaged(ctx context.Context, email, avatarURL, password string) (*User, error) {
	return s.create(ctx, email, avatarURL, password, false)
}

// create validates and persists a local user, optionally enforcing self-registration policy.
func (s *Service) create(ctx context.Context, email, avatarURL, password string, enforceRegistration bool) (*User, error) {
	normalized, err := utils.NormalizeEmail(email)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}
	if enforceRegistration && !s.AllowRegistration() {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorForbidden, "registration is disabled", errorScope)
	}
	password = strings.TrimSpace(password)
	if password == "" {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, "email and password are required", errorScope)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	user, err := s.store.CreateUser(ctx, normalized, avatarURL, string(hash))
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return user, nil
}

// SetLocalPassword enables or replaces a user's local credential and revokes existing sessions.
func (s *Service) SetLocalPassword(ctx context.Context, userID, password string) error {
	userID, err := requireEntityID(userID, "user id is required")
	if err != nil {
		return err
	}
	password = strings.TrimSpace(password)
	if password == "" {
		return errpkg.NewErrorWithScope(errpkg.ErrorValidation, "password is required", errorScope)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	if err := s.store.UpdateUserPassword(ctx, userID, string(hash)); err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return s.store.DeleteUserSessions(ctx, userID)
}

// DisableLocalPassword removes a user's local credential and revokes its sessions.
func (s *Service) DisableLocalPassword(ctx context.Context, userID string) error {
	userID, err := requireEntityID(userID, "user id is required")
	if err != nil {
		return err
	}
	if err := s.store.UpdateUserPassword(ctx, userID, ""); err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return s.store.DeleteUserSessions(ctx, userID)
}

// RevokeSessions invalidates every local browser session for a user.
func (s *Service) RevokeSessions(ctx context.Context, userID string) error {
	userID, err := requireEntityID(userID, "user id is required")
	if err != nil {
		return err
	}
	if err := s.store.DeleteUserSessions(ctx, userID); err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return nil
}

// UpdateRole toggles admin access.
func (s *Service) UpdateRole(ctx context.Context, id string, isAdmin bool) error {
	cleanID, err := requireEntityID(id, "user id is required")
	if err != nil {
		return err
	}
	if err := s.store.UpdateUserAdmin(ctx, cleanID, isAdmin); err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return nil
}

// UpdateName changes the display name for the user profile.
func (s *Service) UpdateName(ctx context.Context, userID, name string) error {
	cleanName := strings.TrimSpace(name)
	if cleanName == "" {
		return errpkg.NewErrorWithScope(errpkg.ErrorValidation, "name is required", errorScope)
	}
	if err := s.store.UpdateUserName(ctx, userID, cleanName); err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return nil
}
