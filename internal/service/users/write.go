package users

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/gi8lino/motus/internal/db"
	errpkg "github.com/gi8lino/motus/internal/service/errors"
	"github.com/gi8lino/motus/internal/utils"
)

// Create registers a new local user.
func (s *Service) Create(ctx context.Context, email, avatarURL, password string) (*User, error) {
	normalized, err := utils.NormalizeEmail(email)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}

	if !s.allowRegistration {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorForbidden, "registration is disabled", errorScope)
	}

	password = strings.TrimSpace(password)
	if password == "" {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, "email and password are required", errorScope)
	}

	var passwordHash string
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
		}
		passwordHash = string(hash)
	}

	user, err := s.store.CreateUser(ctx, normalized, avatarURL, passwordHash)
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

// LinkIdentity connects an administrator-selected external identity to a native user.
func (s *Service) LinkIdentity(ctx context.Context, userID string, identity db.AuthIdentity) error {
	userID, err := requireEntityID(userID, "user id is required")
	if err != nil {
		return err
	}
	if strings.TrimSpace(identity.Provider) == "" || strings.TrimSpace(identity.Subject) == "" {
		return errpkg.NewErrorWithScope(errpkg.ErrorValidation, "provider and subject are required", errorScope)
	}
	if err := s.store.LinkAuthIdentity(ctx, userID, identity); err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return nil
}

// UnlinkIdentity disconnects one external identity from a native user.
func (s *Service) UnlinkIdentity(ctx context.Context, userID, provider, subject string) error {
	userID, err := requireEntityID(userID, "user id is required")
	if err != nil {
		return err
	}
	if err := s.store.DeleteAuthIdentity(ctx, userID, provider, subject); err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return nil
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
