package handler

import (
	"net/http"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service/workouts"
)

// GetWorkouts lists workouts for the current user.
func (a *API) GetWorkouts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		resolvedID, err := a.resolveUserID(r, userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		workouts, err := a.Workouts.List(r.Context(), resolvedID)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, workouts)
	}
}

// CreateWorkout stores a new workout for the current user.
func (a *API) CreateWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		req, err := decode[workouts.WorkoutRequest](r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		resolvedUserID, err := a.resolveUserID(r, userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.UserID = resolvedUserID

		created, err := a.Workouts.Create(r.Context(), req)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, created)
	}
}

// GetWorkout returns a workout by id.
func (a *API) GetWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		workout, err := a.Workouts.Get(r.Context(), id)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, workout)
	}
}

// ExportWorkout returns a workout with nested steps/exercises for sharing.
func (a *API) ExportWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		workout, err := a.Workouts.Export(r.Context(), id)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, workout)
	}
}

// ImportWorkout creates a new workout from exported JSON.
func (a *API) ImportWorkout() http.HandlerFunc {
	type importWorkoutRequest struct {
		UserID  string     `json:"userId"`
		Workout db.Workout `json:"workout"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[importWorkoutRequest](r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		resolvedUserID, err := a.resolveUserID(r, req.UserID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.UserID = resolvedUserID

		created, err := a.Workouts.Import(r.Context(), req.UserID, req.Workout)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, created)
	}
}

// UpdateWorkout replaces a workout and its steps.
func (a *API) UpdateWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		req, err := decode[workouts.WorkoutRequest](r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		if a.AuthHeader != "" {
			resolvedUserID, err := a.resolveUserID(r, "")
			if err != nil {
				writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
				return
			}
			req.UserID = resolvedUserID
		} else if req.UserID == "" {
			current, err := a.Workouts.Get(r.Context(), id)
			if err != nil {
				writeJSON(w, http.StatusNotFound, apiError{Error: err.Error()})
				return
			}
			req.UserID = current.UserID
		}

		updated, err := a.Workouts.Update(r.Context(), id, req)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, updated)
	}
}

// DeleteWorkout removes a workout by id.
func (a *API) DeleteWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := a.Workouts.Delete(r.Context(), id); err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
