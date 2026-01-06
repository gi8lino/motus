package users

import (
	"context"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service"
	"github.com/gi8lino/motus/internal/utils"
)

// store defines the persistence methods needed by the users service.
type store interface {
	CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error)
	UpdateUserAdmin(ctx context.Context, id string, isAdmin bool) error
	GetUserWithPassword(ctx context.Context, id string) (*db.User, string, error)
	UpdateUserPassword(ctx context.Context, id, passwordHash string) error
}

// Service coordinates user operations.
type Service struct {
	Store             store
	AuthHeader        string
	AllowRegistration bool
}

// New creates a new users service.
func New(store store, authHeader string, allowRegistration bool) *Service {
	return &Service{Store: store, AuthHeader: authHeader, AllowRegistration: allowRegistration}
}

// Create registers a new local user.
func (s *Service) Create(ctx context.Context, email, avatarURL, password string) (*db.User, error) {
	normalized, err := utils.NormalizeEmail(email)
	if err != nil {
		return nil, service.NewError(service.ErrorValidation, err.Error())
	}
	password = strings.TrimSpace(password)
	if s.AuthHeader == "" && !s.AllowRegistration {
		return nil, service.NewError(service.ErrorForbidden, "registration is disabled")
	}
	if s.AuthHeader == "" && password == "" {
		return nil, service.NewError(service.ErrorValidation, "password is required")
	}
	passwordHash := ""
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, service.NewError(service.ErrorInternal, "unable to secure password")
		}
		passwordHash = string(hash)
	}
	user, err := s.Store.CreateUser(ctx, normalized, avatarURL, passwordHash)
	if err != nil {
		return nil, service.NewError(service.ErrorInternal, err.Error())
	}
	return user, nil
}

// UpdateRole toggles admin access.
func (s *Service) UpdateRole(ctx context.Context, id string, isAdmin bool) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return service.NewError(service.ErrorValidation, "user id is required")
	}
	if err := s.Store.UpdateUserAdmin(ctx, id, isAdmin); err != nil {
		return service.NewError(service.ErrorInternal, err.Error())
	}
	return nil
}

// Login validates credentials when using local authentication.
func (s *Service) Login(ctx context.Context, email, password string) (*db.User, error) {
	if s.AuthHeader != "" {
		return nil, service.NewError(service.ErrorForbidden, "local login disabled")
	}
	normalized, err := utils.NormalizeEmail(email)
	if err != nil {
		return nil, service.NewError(service.ErrorValidation, err.Error())
	}
	password = strings.TrimSpace(password)
	if password == "" {
		return nil, service.NewError(service.ErrorValidation, "email and password are required")
	}
	user, hash, err := s.Store.GetUserWithPassword(ctx, normalized)
	if err != nil || user == nil {
		return nil, service.NewError(service.ErrorUnauthorized, "invalid credentials")
	}
	if hash == "" {
		return nil, service.NewError(service.ErrorUnauthorized, "password not set")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return nil, service.NewError(service.ErrorUnauthorized, "invalid credentials")
	}
	return user, nil
}

// ChangePassword updates the password for the current user.
func (s *Service) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if s.AuthHeader != "" {
		return service.NewError(service.ErrorForbidden, "passwords managed by proxy")
	}
	currentPassword = strings.TrimSpace(currentPassword)
	newPassword = strings.TrimSpace(newPassword)
	if currentPassword == "" || newPassword == "" {
		return service.NewError(service.ErrorValidation, "current and new password are required")
	}
	_, hash, err := s.Store.GetUserWithPassword(ctx, userID)
	if err != nil || hash == "" {
		return service.NewError(service.ErrorUnauthorized, "invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword)); err != nil {
		return service.NewError(service.ErrorUnauthorized, "invalid credentials")
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return service.NewError(service.ErrorInternal, "unable to secure password")
	}
	if err := s.Store.UpdateUserPassword(ctx, userID, string(newHash)); err != nil {
		return service.NewError(service.ErrorInternal, err.Error())
	}
	return nil
}
