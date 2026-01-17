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
	normalized, err := utils.NormalizeEmail(email)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}

	if s.authHeader == "" && !s.allowRegistration {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorForbidden, "registration is disabled", errorScope)
	}

	password = strings.TrimSpace(password)
	if s.authHeader == "" && password == "" {
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
