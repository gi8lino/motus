package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/logging"
)

// store defines the persistence methods needed by EnsureAdminUser.
type store interface {
	// UpsertAdminUser creates or updates the bootstrap admin user.
	UpsertAdminUser(ctx context.Context, email, passwordHash string) (*db.User, bool, error)
}

// EnsureAdminUser creates or updates the bootstrap admin when configured.
func EnsureAdminUser(ctx context.Context, store store, logger *slog.Logger, email, password string) error {
	email = strings.TrimSpace(email)
	if email == "" && password == "" {
		return nil
	}
	if email == "" || password == "" {
		return fmt.Errorf("admin email and password must both be set")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid admin email: %w", err)
	}

	password = strings.TrimSpace(password)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	user, created, err := store.UpsertAdminUser(ctx, email, string(hash))
	if err != nil {
		return err
	}
	if created {
		logging.SystemLogger(logger, ctx).Info(
			"created bootstrap admin user",
			"event", "bootstrap_admin_created",
			"resource", "user",
			"resource_id", user.ID,
		)
		return nil
	}
	logging.SystemLogger(logger, ctx).Info(
		"updated bootstrap admin user",
		"event", "bootstrap_admin_updated",
		"resource", "user",
		"resource_id", user.ID,
	)

	return nil
}
