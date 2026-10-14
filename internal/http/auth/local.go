package httpauth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	appkitauth "github.com/containeroo/appkit/auth"
	appkitlocal "github.com/containeroo/appkit/auth/local"
	appkitsession "github.com/containeroo/appkit/auth/session"
	"github.com/gi8lino/motus/internal/db"
	"github.com/jackc/pgx/v5"
)

const localSessionCookie = "motus_session"
const sessionHashPurpose = "motus-browser-session-v1"

// CookieOptions configures session-cookie routing and transport security.
type CookieOptions struct {
	PublicURL   string
	RoutePrefix string
}

// Local bridges Motus user persistence with the shared password and session lifecycle.
type Local struct {
	repository localRepository
	sessions   appkitsession.Manager[db.User]
	lifetime   time.Duration
	cookie     appkitsession.CookieConfig
}

// NewLocal constructs a local password and revocable browser session adapter.
func NewLocal(repository localRepository, lifetime time.Duration, opts ...CookieOptions) *Local {
	if lifetime <= 0 {
		lifetime = 30 * 24 * time.Hour
	}
	var option CookieOptions
	if len(opts) != 0 {
		option = opts[0]
	}
	path := "/"
	if option.RoutePrefix != "" {
		path = strings.TrimSuffix(option.RoutePrefix, "/") + "/"
	}
	return &Local{
		repository: repository,
		sessions: appkitsession.Manager[db.User]{
			Store:    sessionStore{repository: repository},
			Lifetime: lifetime,
			Purpose:  sessionHashPurpose,
		},
		lifetime: lifetime,
		cookie: appkitsession.CookieConfig{
			Name:     localSessionCookie,
			Path:     path,
			Lifetime: lifetime,
			Secure:   strings.HasPrefix(option.PublicURL, "https://"),
		},
	}
}

// sessionStore implements the shared session contract using purpose-separated token hashes.
type sessionStore struct{ repository localRepository }

// Create persists a user session by its hash, never its bearer token.
func (s sessionStore) Create(ctx context.Context, hash string, user db.User, expiry time.Time) error {
	return s.repository.CreateSessionHash(ctx, hash, user.ID, expiry)
}

// Lookup resolves a valid session and maps missing rows to shared authentication errors.
func (s sessionStore) Lookup(ctx context.Context, hash string) (db.User, error) {
	user, err := s.repository.GetSessionUserByHash(ctx, hash, time.Now().UTC())
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && user == nil) {
		return db.User{}, appkitsession.ErrNotFound
	}
	if err != nil {
		return db.User{}, err
	}
	return *user, nil
}

// Revoke removes a session hash from durable persistence.
func (s sessionStore) Revoke(ctx context.Context, hash string) error {
	return s.repository.DeleteSessionHash(ctx, hash)
}

// Authenticate resolves a shared session or a valid legacy session during migration.
func (l *Local) Authenticate(r *http.Request) (db.User, error) {
	cookie, err := r.Cookie(localSessionCookie)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return db.User{}, ErrUnauthenticated
	}
	user, err := l.sessions.Resolve(r.Context(), cookie.Value)
	if err == nil || !errors.Is(err, ErrUnauthenticated) {
		return user, err
	}
	// Existing sessions use SHA-256(raw token) rather than appkit's purpose-separated hash.
	previous, err := l.repository.GetSessionUser(r.Context(), cookie.Value, time.Now().UTC())
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && previous == nil) {
		return db.User{}, ErrUnauthenticated
	}
	if err != nil {
		return db.User{}, err
	}
	return *previous, nil
}

// SignIn authenticates existing local credentials and starts an appkit session.
func (l *Local) SignIn(ctx context.Context, email, password string) (db.User, string, error) {
	user, hash, err := l.repository.GetUserWithPassword(ctx, strings.TrimSpace(email))
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && (user == nil || hash == "")) {
		return db.User{}, "", ErrInvalidCredentials
	}
	if err != nil {
		return db.User{}, "", err
	}
	if err := appkitlocal.VerifyPassword(hash, password); err != nil {
		return db.User{}, "", ErrInvalidCredentials
	}
	token, err := l.sessions.Start(ctx, *user)
	if err != nil {
		return db.User{}, "", err
	}
	return *user, token, nil
}

// Setup atomically creates the first administrator and issues a browser session.
func (l *Local) Setup(ctx context.Context, email, password string) (db.User, string, error) {
	if strings.TrimSpace(password) == "" {
		return db.User{}, "", errors.New("password is required")
	}
	hash, err := appkitlocal.HashPassword(password)
	if err != nil {
		return db.User{}, "", err
	}
	user, err := l.repository.CreateInitialAdministrator(ctx, email, hash)
	if err != nil {
		return db.User{}, "", err
	}
	if user == nil {
		return db.User{}, "", errors.New("initial administrator not created")
	}
	token, err := l.sessions.Start(ctx, *user)
	if err != nil {
		return db.User{}, "", err
	}
	return *user, token, nil
}

// StartSession creates a session for an externally verified account.
func (l *Local) StartSession(ctx context.Context, user db.User) (string, error) {
	return l.sessions.Start(ctx, user)
}

// WriteSessionCookie issues a secure browser cookie for a validated session.
func (l *Local) WriteSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	cfg := l.cookie
	cfg.Secure = cfg.Secure || r.TLS != nil
	if cfg.Path != "/" {
		http.SetCookie(w, &http.Cookie{Name: cfg.Name, Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	}
	http.SetCookie(w, cfg.Cookie(token, false))
}

// ClearSession revokes new and legacy session hashes and clears the cookie.
func (l *Local) ClearSession(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(localSessionCookie); err == nil && cookie.Value != "" {
		_ = l.sessions.End(r.Context(), cookie.Value)
		_ = l.repository.DeleteSession(r.Context(), cookie.Value)
	}
	if l.cookie.Path != "/" {
		http.SetCookie(w, &http.Cookie{Name: l.cookie.Name, Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode})
	}
	http.SetCookie(w, l.cookie.Cookie("", true))
}

// RevokeUserSessions invalidates every browser session for the selected user.
func (l *Local) RevokeUserSessions(ctx context.Context, userID string) error {
	return l.repository.DeleteUserSessions(ctx, userID)
}

// ChangePassword checks the old secret, persists its replacement, and revokes sessions.
func (l *Local) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	currentPassword, newPassword = strings.TrimSpace(currentPassword), strings.TrimSpace(newPassword)
	if currentPassword == "" || newPassword == "" {
		return errors.New("currentPassword and newPassword are required")
	}
	user, hash, err := l.repository.GetUserWithPassword(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && (user == nil || user.ID != userID || hash == "")) {
		return ErrInvalidCredentials
	}
	if err != nil {
		return err
	}
	if err := appkitlocal.VerifyPassword(hash, currentPassword); err != nil {
		return ErrInvalidCredentials
	}
	newHash, err := appkitlocal.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := l.repository.UpdateUserPassword(ctx, userID, newHash); err != nil {
		return err
	}
	return l.repository.DeleteUserSessions(ctx, userID)
}

// SetupRequired reports whether local administrator setup is still available.
func (l *Local) SetupRequired(ctx context.Context) (bool, error) {
	return l.repository.SetupRequired(ctx)
}

var _ appkitauth.Authenticator[db.User] = (*Local)(nil)
