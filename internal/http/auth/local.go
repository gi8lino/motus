package httpauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gi8lino/motus/internal/db"
	"golang.org/x/crypto/bcrypt"
)

const localSessionCookie = "motus_session"

// Local owns local credentials and the complete local browser-session lifecycle.
type Local struct {
	repository localRepository
	lifetime   time.Duration
}

func NewLocal(repository localRepository, lifetime time.Duration) *Local {
	if lifetime <= 0 {
		lifetime = 30 * 24 * time.Hour
	}
	return &Local{repository: repository, lifetime: lifetime}
}

func (l *Local) Authenticate(r *http.Request) (db.User, error) {
	cookie, err := r.Cookie(localSessionCookie)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return db.User{}, ErrUnauthenticated
	}
	user, err := l.repository.GetSessionUser(r.Context(), cookie.Value, time.Now().UTC())
	if err != nil || user == nil {
		return db.User{}, ErrUnauthenticated
	}
	return *user, nil
}

// SignIn validates a local credential and creates the browser session.
func (l *Local) SignIn(ctx context.Context, email, password string) (db.User, string, error) {
	user, hash, err := l.repository.GetUserWithPassword(ctx, strings.TrimSpace(email))
	if err != nil || user == nil || hash == "" || bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return db.User{}, "", ErrInvalidCredentials
	}
	token, err := l.createSession(ctx, user.ID)
	if err != nil {
		return db.User{}, "", err
	}
	return *user, token, nil
}

// Setup creates Motus's first local administrator and starts its browser session.
func (l *Local) Setup(ctx context.Context, email, password string) (db.User, string, error) {
	if strings.TrimSpace(password) == "" {
		return db.User{}, "", errors.New("password is required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, "", err
	}
	user, err := l.repository.CreateInitialAdministrator(ctx, email, string(hash))
	if err != nil || user == nil {
		return db.User{}, "", err
	}
	token, err := l.createSession(ctx, user.ID)
	if err != nil {
		return db.User{}, "", err
	}
	return *user, token, nil
}

// StartSession creates a session for a user authenticated by another browser method.
func (l *Local) StartSession(ctx context.Context, user db.User) (string, error) {
	return l.createSession(ctx, user.ID)
}

func (l *Local) createSession(ctx context.Context, userID string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	if err := l.repository.CreateSession(ctx, token, userID, time.Now().UTC().Add(l.lifetime)); err != nil {
		return "", err
	}
	return token, nil
}

// WriteSessionCookie records a local session in the browser.
func (l *Local) WriteSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{Name: localSessionCookie, Value: token, Path: "/", Expires: time.Now().UTC().Add(l.lifetime), HttpOnly: true, Secure: r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https", SameSite: http.SameSiteLaxMode})
}

// ClearSession revokes the current session and clears its browser cookie.
func (l *Local) ClearSession(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(localSessionCookie); err == nil && cookie.Value != "" {
		_ = l.repository.DeleteSession(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: localSessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
}

func (l *Local) RevokeUserSessions(ctx context.Context, userID string) error {
	return l.repository.DeleteUserSessions(ctx, userID)
}

// ChangePassword verifies and replaces a local credential, then revokes every
// existing browser session so the old credential cannot remain active.
func (l *Local) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	currentPassword, newPassword = strings.TrimSpace(currentPassword), strings.TrimSpace(newPassword)
	if currentPassword == "" || newPassword == "" {
		return errors.New("currentPassword and newPassword are required")
	}
	user, hash, err := l.repository.GetUserWithPassword(ctx, userID)
	if err != nil || user == nil || user.ID != userID || hash == "" || bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword)) != nil {
		return ErrInvalidCredentials
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := l.repository.UpdateUserPassword(ctx, userID, string(newHash)); err != nil {
		return err
	}
	return l.repository.DeleteUserSessions(ctx, userID)
}
func (l *Local) SetupRequired(ctx context.Context) (bool, error) {
	return l.repository.SetupRequired(ctx)
}
