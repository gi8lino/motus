package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

// localAuthHeader is the fallback header for local auth.
const localAuthHeader = "X-User-ID"

// store defines the persistence methods needed by auth helpers.
type store interface {
	// GetUser returns a user by id for auth lookups.
	GetUser(ctx context.Context, email string) (*db.User, error)
	// CreateUser inserts a new user for auto-provisioning.
	CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error)
}

// ResolveUserID selects the user id from auth header or request payload.
func ResolveUserID(r *http.Request, store store, authHeader string, autoCreateUsers bool, fallback string) (string, error) {
	// Prefer proxy auth header when configured.
	if authHeader != "" {
		id := strings.TrimSpace(r.Header.Get(authHeader))
		if id == "" {
			return "", fmt.Errorf("auth header is required")
		}

		email, err := utils.NormalizeEmail(id)
		if err != nil {
			return "", err
		}

		// Optionally auto-provision users for new headers.
		if autoCreateUsers {
			if err := ensureUser(r.Context(), store, email); err != nil {
				return "", err
			}
		}
		return email, nil
	}

	if fallback != "" {
		// Accept a fallback id when provided by handlers.
		email, err := utils.NormalizeEmail(fallback)
		if err != nil {
			return "", err
		}
		return email, nil
	}

	// Require the local auth header when no proxy header is configured.
	id := strings.TrimSpace(r.Header.Get(localAuthHeader))
	if id == "" {
		return "", fmt.Errorf("userId is required")
	}

	email, err := utils.NormalizeEmail(id)
	if err != nil {
		return "", err
	}

	return email, nil
}

// ensureUser creates a user if it does not already exist.
func ensureUser(ctx context.Context, store store, email string) error {
	// Short-circuit when the user already exists.
	user, err := store.GetUser(ctx, email)
	if err == nil && user != nil {
		return nil
	}

	// Bubble up unexpected lookup errors.
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	// Create a placeholder user; retry lookup to handle races.
	if _, err := store.CreateUser(ctx, email, "", ""); err != nil {
		// Guard against race conditions if another request created the user.
		if _, getErr := store.GetUser(ctx, email); getErr == nil {
			return nil
		}
		return err
	}

	return nil
}
