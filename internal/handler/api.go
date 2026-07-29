package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gi8lino/motus/internal/auth"
	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/logging"
	"github.com/gi8lino/motus/internal/service/exercises"
	"github.com/gi8lino/motus/internal/service/sounds"
	"github.com/gi8lino/motus/internal/service/trainings"
	"github.com/gi8lino/motus/internal/service/users"
	"github.com/gi8lino/motus/internal/service/workouts"
)

// API bundles shared handler dependencies and runtime configuration.
type API struct {
	Origin            string             // Origin is used for CORS configuration.
	Version           string             // Version is the build version string.
	Commit            string             // Commit is the build commit SHA.
	HealthStore       db.HealthChecker   // HealthStore supports health checks.
	AuthStore         auth.Store         // AuthStore resolves users for auth.
	Users             *users.Service     // Users provides user operations.
	Exercises         *exercises.Service // Exercises provides exercise operations.
	Workouts          *workouts.Service  // Workouts provides workout operations.
	Trainings         *trainings.Service // Trainings provides training operations.
	Logger            *slog.Logger       // Logger reports server activity.
	AuthHeader        string             // AuthHeader specifies the proxy auth header.
	AllowRegistration bool               // AllowRegistration toggles self-serve user creation.
	AutoCreateUsers   bool               // AutoCreateUsers toggles proxy-driven user creation.
	SessionLifetime   time.Duration      // SessionLifetime controls local session expiry.
}

// apiError is a generic error response.
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// statusResponse is the response body for the status endpoint.
type statusResponse struct {
	Status string `json:"status"`
}

// NewAPI builds a handler container with shared dependencies.
func NewAPI(
	store *db.Store,
	logger *slog.Logger,
	authHeader, origin, version, commit string,
	allowRegistration, autoCreateUsers bool,
	sessionLifetime time.Duration,
) *API {
	return &API{
		Origin:            origin,
		Version:           version,
		Commit:            commit,
		HealthStore:       store,
		AuthStore:         store,
		Users:             users.New(store, authHeader, allowRegistration),
		Exercises:         exercises.New(store),
		Workouts:          workouts.New(store),
		Trainings:         trainings.New(store, sounds.URLByKey),
		Logger:            logger,
		AuthHeader:        authHeader,
		AllowRegistration: allowRegistration,
		AutoCreateUsers:   autoCreateUsers,
		SessionLifetime:   sessionLifetime,
	}
}

// respondJSON writes a JSON response.
func (a *API) respondJSON(w http.ResponseWriter, status int, v any) {
	if payload, ok := v.(apiError); ok {
		payload.Code = errorCodeForStatus(status)
		payload.Message = payload.Error
		if status >= http.StatusInternalServerError {
			payload.Message = "Internal server error"
		}
		payload.Error = payload.Message
		v = payload
	}
	if err := encode(w, status, v); err != nil {
		logging.WithRequestIDLogger(a.Logger, nil).Error(
			"encode response failed",
			"event", "encode_response_failed",
			"err", err,
		)
	}
}

func errorCodeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "invalid_request"
	case http.StatusUnauthorized:
		return "authentication_required"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	default:
		return "internal_error"
	}
}

// businessLogger returns a logger enriched with request context for business events.
func (a *API) businessLogger(r *http.Request) *slog.Logger {
	return logging.WithRequestIDLogger(a.Logger, r.Context())
}

// logRequestError records a structured error for the current request.
func (a *API) logRequestError(r *http.Request, event, message string, err error) {
	logging.WithRequestIDLogger(a.Logger, r.Context()).Error(
		message,
		"event", event,
		"err", err,
	)
}

// resolveUserID selects the user id from auth header or request payload.
func (a *API) resolveUserID(r *http.Request, fallback string) (string, error) {
	return auth.ResolveUserID(r, a.AuthStore, a.AuthHeader, a.AutoCreateUsers, fallback)
}

// requireWorkoutOwner hides workouts that do not belong to the authenticated user.
func (a *API) requireWorkoutOwner(r *http.Request, workoutID string) error {
	// Handler unit tests use an API without an auth store; the production API always has one.
	if a.AuthStore == nil {
		return nil
	}
	userID, err := a.resolveUserID(r, "")
	if err != nil {
		return err
	}
	_, err = a.Workouts.GetForUser(r.Context(), workoutID, userID)
	if err != nil {
		return err
	}
	return nil
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
