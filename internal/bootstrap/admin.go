package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/gi8lino/motus/internal/db"
)

// store defines the persistence methods needed by EnsureAdminUser.
type store interface {
	UpsertAdminUser(ctx context.Context, email, passwordHash string) (*db.User, bool, error)
}

// EnsureAdminUser creates or updates the bootstrap admin when configured.
func EnsureAdminUser(ctx context.Context, store store, logger *slog.Logger, email, password string) error {
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" && password == "" {
		return nil
	}
	if email == "" || password == "" {
		return fmt.Errorf("admin email and password must both be set")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid admin email: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	user, created, err := store.UpsertAdminUser(ctx, email, string(hash))
	if err != nil {
		return err
	}
	if created {
		logger.Info("created bootstrap admin user", "user", user.ID)
		return nil
	}
	logger.Info("updated bootstrap admin user", "user", user.ID)
	return nil
}
