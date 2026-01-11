package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gi8lino/motus/internal/auth"
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service/exercises"
	"github.com/gi8lino/motus/internal/service/sessions"
	"github.com/gi8lino/motus/internal/service/templates"
	"github.com/gi8lino/motus/internal/service/users"
	"github.com/gi8lino/motus/internal/service/workouts"
)

// API bundles shared handler dependencies and runtime configuration.
type API struct {
	Origin            string           // Origin is used for CORS configuration.
	Version           string           // Version is the build version string.
	Commit            string           // Commit is the build commit SHA.
	HealthStore       db.HealthChecker // HealthStore supports health checks.
	AuthStore         auth.Store       // AuthStore resolves users for auth.
	UsersStore        users.Store      // Users provides user persistence.
	ExercisesStore    exercises.Store  // Exercises provides exercise persistence.
	WorkoutsStore     workouts.Store   // Workouts provides workout persistence.
	TemplatesStore    templates.Store  // Templates provides template persistence.
	SessionsStore     sessions.Store   // Sessions provides session persistence.
	Logger            *slog.Logger     // Logger reports server activity.
	AuthHeader        string           // AuthHeader specifies the proxy auth header.
	AllowRegistration bool             // AllowRegistration toggles self-serve user creation.
	AutoCreateUsers   bool             // AutoCreateUsers toggles proxy-driven user creation.
}

// apiError is a generic error response.
type apiError struct {
	Error string `json:"error"`
}

// NewAPI builds a handler container with shared dependencies.
func NewAPI(store *db.Store, logger *slog.Logger, authHeader, origin, version, commit string, allowRegistration, autoCreateUsers bool) *API {
	return &API{
		Origin:            origin,
		Version:           version,
		Commit:            commit,
		HealthStore:       store,
		AuthStore:         store,
		UsersStore:        store,
		ExercisesStore:    store,
		WorkoutsStore:     store,
		TemplatesStore:    store,
		SessionsStore:     store,
		Logger:            logger,
		AuthHeader:        authHeader,
		AllowRegistration: allowRegistration,
		AutoCreateUsers:   autoCreateUsers,
	}
}

// resolveUserID selects the user id from auth header or request payload.
func (a *API) resolveUserID(r *http.Request, fallback string) (string, error) {
	return auth.ResolveUserID(r, a.AuthStore, a.AuthHeader, a.AutoCreateUsers, fallback)
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
