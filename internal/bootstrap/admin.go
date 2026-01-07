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
	// Exit early when no bootstrap credentials are configured.
	if email == "" && password == "" {
		return nil
	}
	// Guard: require both fields together to avoid partial setup.
	if email == "" || password == "" {
		return fmt.Errorf("admin email and password must both be set")
	}
	// Validate the email format before hashing/storing.
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid admin email: %w", err)
	}
	// Hash the password so no plaintext is stored.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	// Create or update the admin user in a single upsert.
	user, created, err := store.UpsertAdminUser(ctx, email, string(hash))
	if err != nil {
		return err
	}
	// Log whether a new admin was created or updated.
	if created {
		logger.Info("created bootstrap admin user", "user", user.ID)
		return nil
	}
	logger.Info("updated bootstrap admin user", "user", user.ID)
	return nil
}
