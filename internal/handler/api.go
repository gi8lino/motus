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
	// Ping verifies database connectivity.
	Ping(ctx context.Context) error
	// GetUser returns a user by id.
	GetUser(ctx context.Context, id string) (*db.User, error)
	// ListUsers returns all user records.
	ListUsers(ctx context.Context) ([]db.User, error)
	// GetUserWithPassword returns a user with its password hash.
	GetUserWithPassword(ctx context.Context, id string) (*db.User, string, error)
	// UpdateUserPassword updates the password hash for a user.
	UpdateUserPassword(ctx context.Context, id, passwordHash string) error
	// UpdateUserAdmin sets the admin flag for a user.
	UpdateUserAdmin(ctx context.Context, id string, isAdmin bool) error
	// UpdateUserName updates the display name for a user.
	UpdateUserName(ctx context.Context, id, name string) error
	// CreateUser creates a new user record.
	CreateUser(ctx context.Context, email, avatarURL, passwordHash string) (*db.User, error)
	// ListExercises returns exercises visible to a user.
	ListExercises(ctx context.Context, userID string) ([]db.Exercise, error)
	// CreateExercise adds a new exercise to the catalog.
	CreateExercise(ctx context.Context, name, userID string, isCore bool) (*db.Exercise, error)
	// GetExercise returns a catalog entry by id.
	GetExercise(ctx context.Context, id string) (*db.Exercise, error)
	// RenameExercise updates the name of a catalog entry.
	RenameExercise(ctx context.Context, id, name string) (*db.Exercise, error)
	// ReplaceExerciseForUser clones a core exercise for a user.
	ReplaceExerciseForUser(ctx context.Context, userID, oldID, newID, newName string) error
	// DeleteExercise removes a catalog entry.
	DeleteExercise(ctx context.Context, id string) error
	// BackfillCoreExercises migrates exercises into the core catalog.
	BackfillCoreExercises(ctx context.Context) error
	// ListTemplates returns shared workout templates.
	ListTemplates(ctx context.Context) ([]db.Workout, error)
	// CreateTemplateFromWorkout converts a workout into a template.
	CreateTemplateFromWorkout(ctx context.Context, workoutID, name string) (*db.Workout, error)
	// CreateWorkoutFromTemplate clones a template for a user.
	CreateWorkoutFromTemplate(ctx context.Context, templateID, userID, name string) (*db.Workout, error)
	// WorkoutsByUser returns workouts owned by a user.
	WorkoutsByUser(ctx context.Context, userID string) ([]db.Workout, error)
	// CreateWorkout inserts a workout definition.
	CreateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error)
	// UpdateWorkout updates a workout definition.
	UpdateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error)
	// WorkoutWithSteps loads a workout with its steps.
	WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error)
	// DeleteWorkout removes a workout by id.
	DeleteWorkout(ctx context.Context, id string) error
	// SessionHistory returns completed session logs for a user.
	SessionHistory(ctx context.Context, userID string, limit int) ([]db.SessionLog, error)
	// SessionStepTimings returns step logs for a session.
	SessionStepTimings(ctx context.Context, sessionID string) ([]db.SessionStepLog, error)
	// RecordSession persists a completed session and its steps.
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
