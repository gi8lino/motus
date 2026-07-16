package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

// localAuthHeader is the fallback header for local auth.
const localAuthHeader = "X-User-ID"

const SessionCookieName = "motus_session"
const sessionLifetime = 30 * 24 * time.Hour

type principalKey struct{}

func WithPrincipal(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, principalKey{}, userID)
}

func Principal(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(principalKey{}).(string)
	return id, ok && id != ""
}

// Store defines the persistence methods needed by auth helpers.
type Store interface {
	// GetUser returns a user by id for auth lookups.
	GetUser(ctx context.Context, email string) (*db.User, error)
	// CreateUser inserts a new user for auto-provisioning.
	CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error)
	CreateSession(ctx context.Context, token, userID string, expiresAt time.Time) error
	GetSessionUser(ctx context.Context, token string, now time.Time) (*db.User, error)
	DeleteSession(ctx context.Context, token string) error
	DeleteUserSessions(ctx context.Context, userID string) error
}

// EndSession revokes the current session and expires its browser cookie.
func EndSession(ctx context.Context, w http.ResponseWriter, r *http.Request, store Store) error {
	if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
		if err := store.DeleteSession(ctx, cookie.Value); err != nil {
			return err
		}
	}
	http.SetCookie(w, &http.Cookie{Name: SessionCookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	return nil
}

// StartSession creates a local-auth session and writes its opaque token cookie.
func StartSession(ctx context.Context, w http.ResponseWriter, r *http.Request, store Store, userID string, lifetime time.Duration) error {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	if lifetime <= 0 {
		lifetime = sessionLifetime
	}
	expires := time.Now().UTC().Add(lifetime)
	if err := store.CreateSession(ctx, token, userID, expires); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name: SessionCookieName, Value: token, Path: "/", Expires: expires,
		HttpOnly: true, Secure: r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// ResolveUserID selects the user id from auth header or request payload.
func ResolveUserID(r *http.Request, store Store, authHeader string, autoCreateUsers bool, fallback string) (string, error) {
	if id, ok := Principal(r.Context()); ok {
		return id, nil
	}
	// Prefer proxy auth header when configured.
	if authHeader != "" {
		id := strings.TrimSpace(r.Header.Get(authHeader))
		if id == "" {
			return "", errors.New("auth header is required")
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
	// Handler unit tests may omit the auth store; production always supplies one.
	if store == nil && fallback != "" {
		return utils.NormalizeEmail(fallback)
	}
	if store == nil {
		return utils.NormalizeEmail(r.Header.Get(localAuthHeader))
	}

	// Local authentication uses an opaque, HTTP-only session cookie.
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return "", errors.New("authentication required")
	}
	user, err := store.GetSessionUser(r.Context(), cookie.Value, time.Now().UTC())
	if err != nil || user == nil {
		return "", errors.New("invalid or expired session")
	}
	return user.ID, nil
}

// ensureUser creates a user if it does not already exist.
func ensureUser(ctx context.Context, store Store, email string) error {
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
