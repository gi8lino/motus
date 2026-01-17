package users

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
	"github.com/gi8lino/motus/internal/utils"
)

// Login validates credentials when using local authentication.
func (s *Service) Login(ctx context.Context, email, password string) (*User, error) {
	if s.authHeader != "" {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorForbidden, "passwords managed by proxy", errorScope)
	}

	normalized, err := utils.NormalizeEmail(email)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, err.Error(), errorScope)
	}

	password = strings.TrimSpace(password)
	if password == "" {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, "email and password are required", errorScope)
	}

	user, hash, err := s.store.GetUserWithPassword(ctx, normalized)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	if user == nil || hash == "" {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorUnauthorized, "invalid credentials", errorScope)
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorUnauthorized, "invalid credentials", errorScope)
	}

	return user, nil
}

// ChangePassword updates the password for the current user.
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if s.authHeader != "" {
		return errpkg.NewErrorWithScope(errpkg.ErrorForbidden, "passwords managed by proxy", errorScope)
	}

	currentPassword = strings.TrimSpace(currentPassword)
	newPassword = strings.TrimSpace(newPassword)
	if currentPassword == "" || newPassword == "" {
		return errpkg.NewErrorWithScope(errpkg.ErrorValidation, "currentPassword and newPassword are required", errorScope)
	}

	user, hash, err := s.store.GetUserWithPassword(ctx, userID)
	if err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	if user == nil || hash == "" {
		return errpkg.NewErrorWithScope(errpkg.ErrorUnauthorized, "invalid credentials", errorScope)
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword)) != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorUnauthorized, "invalid credentials", errorScope)
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}

	if err := s.store.UpdateUserPassword(ctx, userID, string(newHash)); err != nil {
		return errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return nil
}
