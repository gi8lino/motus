package server

import (
	"net/http"

	"github.com/gi8lino/motus/internal/http/endpoint"
)

// registerBrowserRoutes registers authenticated non-administrator API routes.
func registerBrowserRoutes(mux *http.ServeMux, config Config) {
	mux.Handle("GET /me", endpoint.CurrentUser())
	mux.Handle("POST /logout", endpoint.Logout(config.Browser.Local, config.Logger))
	mux.Handle("PUT /me/password", endpoint.ChangePassword(config.Browser.Local, config.Logger))
	mux.Handle("PUT /me/name", endpoint.UpdateUserName(config.Users, config.Logger))
	mux.Handle("GET /users/{id}/workouts", endpoint.GetWorkouts(config.Workouts, config.Logger))
	mux.Handle("POST /users/{id}/workouts", endpoint.CreateWorkout(config.Workouts, config.Logger))
	mux.Handle("GET /workouts/{id}", endpoint.GetWorkout(config.Workouts, config.Logger))
	mux.Handle("GET /workouts/{id}/export", endpoint.ExportWorkout(config.Workouts, config.Logger))
	mux.Handle("POST /workouts/import", endpoint.ImportWorkout(config.Workouts, config.Logger))
	mux.Handle("PUT /workouts/{id}", endpoint.UpdateWorkout(config.Workouts, config.Logger))
	mux.Handle("DELETE /workouts/{id}", endpoint.DeleteWorkout(config.Workouts, config.Logger))
	mux.Handle("GET /exercises", endpoint.ListExercises(config.Exercises, config.Logger))
	mux.Handle("POST /exercises", endpoint.CreateExercise(config.Exercises, config.Logger))
	mux.Handle("PUT /exercises/{id}", endpoint.UpdateExercise(config.Exercises, config.Logger))
	mux.Handle("DELETE /exercises/{id}", endpoint.DeleteExercise(config.Exercises, config.Logger))
	mux.Handle("GET /sounds", endpoint.ListSounds(config.Logger))
	mux.Handle("POST /trainings", endpoint.CreateTraining(config.Trainings, config.Workouts, config.Logger))
	mux.Handle("GET /users/{id}/trainings/history", endpoint.ListTrainingHistory(config.Trainings, config.Workouts, config.Logger))
	mux.Handle("POST /trainings/complete", endpoint.CompleteTraining(config.Trainings, config.Workouts, config.Logger))
	mux.Handle("PUT /trainings/{id}/feedback", endpoint.UpdateTrainingFeedback(config.Trainings, config.Workouts, config.Logger))
}
