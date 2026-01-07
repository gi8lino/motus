package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gi8lino/motus/internal/auth"
	"github.com/gi8lino/motus/internal/db"
)

// Store defines the persistence methods needed by handlers and services.
type Store interface {
	Ping(ctx context.Context) error
	GetUser(ctx context.Context, id string) (*db.User, error)
	ListUsers(ctx context.Context) ([]db.User, error)
	GetUserWithPassword(ctx context.Context, id string) (*db.User, string, error)
	UpdateUserPassword(ctx context.Context, id, passwordHash string) error
	UpdateUserAdmin(ctx context.Context, id string, isAdmin bool) error
	UpdateUserName(ctx context.Context, id, name string) error
	CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error)
	ListExercises(ctx context.Context, userID string) ([]db.Exercise, error)
	CreateExercise(ctx context.Context, name, userID string, isCore bool) (*db.Exercise, error)
	GetExercise(ctx context.Context, id string) (*db.Exercise, error)
	RenameExercise(ctx context.Context, id, name string) (*db.Exercise, error)
	ReplaceExerciseForUser(ctx context.Context, userID, oldID, newID, newName string) error
	DeleteExercise(ctx context.Context, id string) error
	BackfillCoreExercises(ctx context.Context) error
	ListTemplates(ctx context.Context) ([]db.Workout, error)
	CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*db.Workout, error)
	CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*db.Workout, error)
	WorkoutsByUser(ctx context.Context, userID string) ([]db.Workout, error)
	CreateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error)
	UpdateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error)
	WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error)
	DeleteWorkout(ctx context.Context, id string) error
	SessionHistory(ctx context.Context, userID string, limit int) ([]db.SessionLog, error)
	SessionStepTimings(ctx context.Context, sessionID string) ([]db.SessionStepLog, error)
	RecordSession(ctx context.Context, log db.SessionLog, steps []db.SessionStepLog) error
}

// API bundles shared handler dependencies and runtime configuration.
type API struct {
	Origin            string       // Origin is used for CORS configuration.
	Version           string       // Version is the build version string.
	Commit            string       // Commit is the build commit SHA.
	Store             Store        // Store provides database access.
	Logger            *slog.Logger // Logger reports server activity.
	AuthHeader        string       // AuthHeader specifies the proxy auth header.
	AllowRegistration bool         // AllowRegistration toggles self-serve user creation.
	AutoCreateUsers   bool         // AutoCreateUsers toggles proxy-driven user creation.
}

// apiError is a generic error response.
type apiError struct {
	Error string `json:"error"`
}

// NewAPI builds a handler container with shared dependencies.
func NewAPI(store Store, logger *slog.Logger, authHeader, origin, version, commit string, allowRegistration, autoCreateUsers bool) *API {
	return &API{
		Origin:            origin,
		Version:           version,
		Commit:            commit,
		Store:             store,
		Logger:            logger,
		AuthHeader:        authHeader,
		AllowRegistration: allowRegistration,
		AutoCreateUsers:   autoCreateUsers,
	}
}

// resolveUserID selects the user id from auth header or request payload.
func (a *API) resolveUserID(r *http.Request, fallback string) (string, error) {
	return auth.ResolveUserID(r, a.Store, a.AuthHeader, a.AutoCreateUsers, fallback)
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
