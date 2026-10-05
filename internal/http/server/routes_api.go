package server

import (
	"net/http"

	"github.com/gi8lino/motus/internal/http/endpoint"
	"github.com/gi8lino/motus/internal/http/middleware"
)

// registerAPIRoutes registers authenticated Motus API routes.
func registerAPIRoutes(mux *http.ServeMux, config Config) {
	routes := http.NewServeMux()
	routes.Handle("GET /me", endpoint.CurrentUser())
	routes.Handle("POST /logout", endpoint.Logout(config.Browser.Local, config.Logger))
	routes.Handle("PUT /me/password", endpoint.ChangePassword(config.Browser.Local, config.Logger))
	routes.Handle("PUT /me/name", endpoint.UpdateUserName(config.Users, config.Logger))
	routes.Handle("GET /users", middleware.Chain(endpoint.GetUsers(config.Users, config.Logger), middleware.RequireAdmin(config.Users)))
	routes.Handle("PUT /users/{id}/admin", middleware.Chain(endpoint.UpdateUserRole(config.Users, config.Logger), middleware.RequireAdmin(config.Users)))
	routes.Handle("GET /users/{id}/workouts", endpoint.GetWorkouts(config.Workouts, config.Logger))
	routes.Handle("POST /users/{id}/workouts", endpoint.CreateWorkout(config.Workouts, config.Logger))
	routes.Handle("GET /workouts/{id}", endpoint.GetWorkout(config.Workouts, config.Logger))
	routes.Handle("GET /workouts/{id}/export", endpoint.ExportWorkout(config.Workouts, config.Logger))
	routes.Handle("POST /workouts/import", endpoint.ImportWorkout(config.Workouts, config.Logger))
	routes.Handle("PUT /workouts/{id}", endpoint.UpdateWorkout(config.Workouts, config.Logger))
	routes.Handle("DELETE /workouts/{id}", endpoint.DeleteWorkout(config.Workouts, config.Logger))
	routes.Handle("GET /exercises", endpoint.ListExercises(config.Exercises, config.Logger))
	routes.Handle("POST /exercises", endpoint.CreateExercise(config.Exercises, config.Logger))
	routes.Handle("PUT /exercises/{id}", endpoint.UpdateExercise(config.Exercises, config.Logger))
	routes.Handle("DELETE /exercises/{id}", endpoint.DeleteExercise(config.Exercises, config.Logger))
	routes.Handle("GET /sounds", endpoint.ListSounds(config.Logger))
	routes.Handle("POST /trainings", endpoint.CreateTraining(config.Trainings, config.Workouts, config.Logger))
	routes.Handle("GET /users/{id}/trainings/history", endpoint.ListTrainingHistory(config.Trainings, config.Workouts, config.Logger))
	routes.Handle("POST /trainings/complete", endpoint.CompleteTraining(config.Trainings, config.Workouts, config.Logger))
	routes.Handle("PUT /trainings/{id}/feedback", endpoint.UpdateTrainingFeedback(config.Trainings, config.Workouts, config.Logger))
	mux.Handle("/api/", http.StripPrefix("/api", middleware.Chain(routes, middleware.Authenticate(config.Browser.Authenticator))))
}
