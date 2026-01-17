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
		return nil, errpkg.NewError(errpkg.ErrorForbidden, "passwords managed by proxy")
	}

	normalized, err := utils.NormalizeEmail(email)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorValidation, err.Error())
	}

	password = strings.TrimSpace(password)
	if password == "" {
		return nil, errpkg.NewError(errpkg.ErrorValidation, "email and password are required")
	}

	user, hash, err := s.store.GetUserWithPassword(ctx, normalized)
	if err != nil {
		return nil, errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	if user == nil || hash == "" {
		return nil, errpkg.NewError(errpkg.ErrorUnauthorized, "invalid credentials")
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return nil, errpkg.NewError(errpkg.ErrorUnauthorized, "invalid credentials")
	}

	return user, nil
}

// ChangePassword updates the password for the current user.
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if s.authHeader != "" {
		return errpkg.NewError(errpkg.ErrorForbidden, "passwords managed by proxy")
	}

	currentPassword = strings.TrimSpace(currentPassword)
	newPassword = strings.TrimSpace(newPassword)
	if currentPassword == "" || newPassword == "" {
		return errpkg.NewError(errpkg.ErrorValidation, "currentPassword and newPassword are required")
	}

	user, hash, err := s.store.GetUserWithPassword(ctx, userID)
	if err != nil {
		return errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	if user == nil || hash == "" {
		return errpkg.NewError(errpkg.ErrorUnauthorized, "invalid credentials")
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword)) != nil {
		return errpkg.NewError(errpkg.ErrorUnauthorized, "invalid credentials")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}

	if err := s.store.UpdateUserPassword(ctx, userID, string(newHash)); err != nil {
		return errpkg.NewError(errpkg.ErrorInternal, err.Error())
	}
	return nil
}
