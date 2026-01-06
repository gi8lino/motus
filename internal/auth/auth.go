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
	GetUser(ctx context.Context, email string) (*db.User, error)
	CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error)
}

// ResolveUserID selects the user id from auth header or request payload.
func ResolveUserID(r *http.Request, store store, authHeader string, autoCreateUsers bool, fallback string) (string, error) {
	if authHeader != "" {
		id := strings.TrimSpace(r.Header.Get(authHeader))
		if id == "" {
			return "", fmt.Errorf("auth header is required")
		}
		email, err := utils.NormalizeEmail(id)
		if err != nil {
			return "", err
		}
		if autoCreateUsers {
			if err := ensureUser(r.Context(), store, email); err != nil {
				return "", err
			}
		}
		return email, nil
	}
	if fallback != "" {
		email, err := utils.NormalizeEmail(fallback)
		if err != nil {
			return "", err
		}
		return email, nil
	}
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
	user, err := store.GetUser(ctx, email)
	if err == nil && user != nil {
		return nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if _, err := store.CreateUser(ctx, email, "", ""); err != nil {
		// Guard against race conditions if another request created the user.
		if _, getErr := store.GetUser(ctx, email); getErr == nil {
			return nil
		}
		return err
	}
	return nil
}
