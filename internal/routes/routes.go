package routes

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/gi8lino/motus/internal/handler"
	"github.com/gi8lino/motus/internal/middleware"
)

// NewRouter wires HTTP routes for Motus.
func NewRouter(
	spaFS fs.FS,
	routePrefix string,
	logger *slog.Logger,
	api *handler.API,
	debug bool,
) (http.Handler, error) {
	mux := http.NewServeMux()

	spaContent, err := fs.Sub(spaFS, "web/dist")
	if err != nil {
		return nil, fmt.Errorf("spa filesystem: %w", err)
	}
	spaFiles := http.FileServer(http.FS(spaContent))
	mux.Handle("GET /assets/", spaFiles)
	mux.Handle("GET /fonts/", spaFiles)
	mux.Handle("GET /sounds/", spaFiles)
	mux.Handle("GET /index.html", spaFiles)
	mux.Handle("GET /motus.svg", spaFiles)
	mux.Handle("GET /brand.svg", spaFiles)
	mux.Handle("GET /favicon-16x16.png", spaFiles)
	mux.Handle("GET /favicon-32x32.png", spaFiles)
	mux.Handle("GET /favicon-48x48.png", spaFiles)
	mux.Handle("GET /favicon-64x64.png", spaFiles)
	mux.Handle("GET /apple-touch-icon.png", spaFiles)

	// Healthz
	mux.Handle("GET /healthz", api.Healthz())
	mux.Handle("POST /healthz", api.Healthz())

	// API
	apiMux := http.NewServeMux()
	apiMux.Handle("GET /config", api.Config())
	apiMux.Handle("GET /me", api.CurrentUser())
	apiMux.Handle("POST /login", api.Login())
	apiMux.Handle("PUT /me/password", api.ChangePassword())
	apiMux.Handle("PUT /me/name", api.UpdateUserName())
	apiMux.Handle("GET /users",
		middleware.Chain(api.GetUsers(), middleware.RequireAdmin(api.Users, api.AuthHeader)),
	)
	apiMux.Handle("POST /users", api.CreateUser())
	apiMux.Handle("PUT /users/{id}/admin",
		middleware.Chain(api.UpdateUserRole(),
			middleware.RequireAdmin(api.Users, api.AuthHeader),
		),
	)

	apiMux.Handle("GET /users/{id}/workouts", api.GetWorkouts())
	apiMux.Handle("POST /users/{id}/workouts", api.CreateWorkout())
	apiMux.Handle("GET /workouts/{id}", api.GetWorkout())
	apiMux.Handle("GET /workouts/{id}/export", api.ExportWorkout())
	apiMux.Handle("POST /workouts/import", api.ImportWorkout())
	apiMux.Handle("PUT /workouts/{id}", api.UpdateWorkout())
	apiMux.Handle("DELETE /workouts/{id}", api.DeleteWorkout())

	apiMux.Handle("GET /templates", api.ListTemplates())
	apiMux.Handle("POST /templates", api.CreateTemplate())
	apiMux.Handle("GET /templates/{id}", api.GetTemplate())
	apiMux.Handle("POST /templates/{id}/apply", api.ApplyTemplate())

	apiMux.Handle("GET /exercises", api.ListExercises())
	apiMux.Handle("POST /exercises", api.CreateExercise())
	apiMux.Handle("PUT /exercises/{id}", api.UpdateExercise())
	apiMux.Handle("DELETE /exercises/{id}", api.DeleteExercise())
	apiMux.Handle("POST /exercises/backfill",
		middleware.Chain(
			api.BackfillExercises(),
			middleware.RequireAdmin(api.Users, api.AuthHeader),
		),
	)

	apiMux.Handle("GET /sounds", api.ListSounds())

	apiMux.Handle("POST /trainings", api.CreateTraining())
	apiMux.Handle("GET /users/{id}/trainings/history", api.ListTrainingHistory())
	apiMux.Handle("POST /trainings/complete", api.CompleteTraining())

	// Mount API under /api
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// SPA
	mux.Handle("/", handler.SPA(spaContent, routePrefix))

	var h http.Handler = mux
	h = handler.WithCORS(api.Origin, h)
	if routePrefix != "" {
		logger.Info(
			"mounted under prefix",
			"event", "routes_mounted",
			"prefix", routePrefix,
		)
		h = mountUnderPrefix(h, routePrefix)
	}

	// Optional debug logging middleware.
	if debug {
		return middleware.Chain(h, middleware.LoggingMiddleware(api.Logger), middleware.RequestIDMiddleware()), nil
	}

	return middleware.Chain(h, middleware.RequestIDMiddleware()), nil
}
