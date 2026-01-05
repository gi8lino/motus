package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gi8lino/motus/internal/db"
)

// API bundles shared handler dependencies and runtime configuration.
type API struct {
	Origin            string       // Origin is used for CORS configuration.
	Version           string       // Version is the build version string.
	Commit            string       // Commit is the build commit SHA.
	Store             *db.Store    // Store provides database access.
	Logger            *slog.Logger // Logger reports server activity.
	AuthHeader        string       // AuthHeader specifies the proxy auth header.
	AllowRegistration bool         // AllowRegistration toggles self-serve user creation.
}

// apiError is a generic error response.
type apiError struct {
	Error string `json:"error"`
}

// localAuthHeader is the fallback header for local auth.
const localAuthHeader = "X-User-ID"

// NewAPI builds a handler container with shared dependencies.
func NewAPI(store *db.Store, logger *slog.Logger, authHeader, origin, version, commit string, allowRegistration bool) *API {
	return &API{
		Origin:            origin,
		Version:           version,
		Commit:            commit,
		Store:             store,
		Logger:            logger,
		AuthHeader:        authHeader,
		AllowRegistration: allowRegistration,
	}
}

// resolveUserID selects the user id from auth header or request payload.
func (a *API) resolveUserID(r *http.Request, fallback string) (string, error) {
	if a.AuthHeader != "" {
		id := strings.TrimSpace(r.Header.Get(a.AuthHeader))
		if id == "" {
			return "", fmt.Errorf("auth header is required")
		}
		email, err := normalizeEmail(id)
		if err != nil {
			return "", err
		}
		return email, nil
	}
	if fallback != "" {
		email, err := normalizeEmail(fallback)
		if err != nil {
			return "", err
		}
		return email, nil
	}
	id := strings.TrimSpace(r.Header.Get(localAuthHeader))
	if id == "" {
		return "", fmt.Errorf("userId is required")
	}
	email, err := normalizeEmail(id)
	if err != nil {
		return "", err
	}
	return email, nil
}

// WithCORS adds CORS headers to the handler.
func WithCORS(origin string, next http.Handler) http.Handler {
	if origin == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// writeJSON writes the given payload as JSON to the response.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
