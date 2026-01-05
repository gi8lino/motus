package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/gi8lino/motus/internal/db"
)

// GetWorkouts lists workouts for the current user.
func (a *API) GetWorkouts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := strings.TrimSpace(r.PathValue("id"))
		resolvedID, err := a.resolveUserID(r, userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		workouts, err := a.Store.WorkoutsByUser(r.Context(), resolvedID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, workouts)
	}
}

// CreateWorkout stores a new workout for the current user.
func (a *API) CreateWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := strings.TrimSpace(r.PathValue("id"))
		var req workoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		resolvedUserID, err := a.resolveUserID(r, userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.UserID = resolvedUserID
		if req.UserID == "" || req.Name == "" || len(req.Steps) == 0 {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "name and at least one step are required"})
			return
		}
		steps, err := normalizeSteps(req.Steps)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		workout := &db.Workout{UserID: req.UserID, Name: req.Name, Steps: steps}
		created, err := a.Store.CreateWorkout(r.Context(), workout)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

// GetWorkout returns a workout by id.
func (a *API) GetWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "workout id is required"})
			return
		}
		workout, err := a.Store.WorkoutWithSteps(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, workout)
	}
}

// ExportWorkout returns a workout with nested steps/exercises for sharing.
func (a *API) ExportWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "workout id is required"})
			return
		}
		workout, err := a.Store.WorkoutWithSteps(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, workout)
	}
}

// ImportWorkout creates a new workout from exported JSON.
func (a *API) ImportWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			UserID  string     `json:"userId"`
			Workout db.Workout `json:"workout"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.UserID = strings.TrimSpace(req.UserID)
		resolvedUserID, err := a.resolveUserID(r, req.UserID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.UserID = resolvedUserID
		req.Workout.Name = strings.TrimSpace(req.Workout.Name)
		if req.Workout.Name == "" || len(req.Workout.Steps) == 0 {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "workout name and steps are required"})
			return
		}
		for idx := range req.Workout.Steps {
			step := &req.Workout.Steps[idx]
			step.ID = ""
			step.WorkoutID = ""
			step.Order = idx
			for exIdx := range step.Exercises {
				ex := &step.Exercises[exIdx]
				ex.ID = ""
				ex.StepID = ""
				ex.Order = exIdx
				ex.ExerciseID = ""
			}
		}
		created, err := a.Store.CreateWorkout(r.Context(), &db.Workout{
			UserID: req.UserID,
			Name:   req.Workout.Name,
			Steps:  req.Workout.Steps,
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, created)
	}
}

// UpdateWorkout replaces a workout and its steps.
func (a *API) UpdateWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "workout id is required"})
			return
		}
		var req workoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" || len(req.Steps) == 0 {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "name and steps are required"})
			return
		}
		req.UserID = strings.TrimSpace(req.UserID)
		if a.AuthHeader != "" {
			resolvedUserID, err := a.resolveUserID(r, "")
			if err != nil {
				writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
				return
			}
			req.UserID = resolvedUserID
		} else if req.UserID == "" {
			current, err := a.Store.WorkoutWithSteps(r.Context(), id)
			if err != nil {
				writeJSON(w, http.StatusNotFound, apiError{Error: err.Error()})
				return
			}
			req.UserID = current.UserID
		}
		steps, err := normalizeSteps(req.Steps)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		workout := &db.Workout{ID: id, UserID: req.UserID, Name: req.Name, Steps: steps}
		updated, err := a.Store.UpdateWorkout(r.Context(), workout)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

// DeleteWorkout removes a workout by id.
func (a *API) DeleteWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "workout id is required"})
			return
		}
		if err := a.Store.DeleteWorkout(r.Context(), id); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				writeJSON(w, http.StatusNotFound, apiError{Error: "workout not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
