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
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout list user", "err", err)
			}
			return
		}

		workouts, err := a.Workouts.List(r.Context(), resolvedID)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout list", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, workouts); err != nil {
			a.Logger.Error("workout list encode", "err", err)
		}
	}
}

// CreateWorkout stores a new workout for the current user.
func (a *API) CreateWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		req, err := decode[workouts.WorkoutRequest](r)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout create decode", "err", err)
			}
			return
		}

		resolvedUserID, err := a.resolveUserID(r, userID)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout create user", "err", err)
			}
			return
		}
		req.UserID = resolvedUserID

		created, err := a.Workouts.Create(r.Context(), req)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout create", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusCreated, created); err != nil {
			a.Logger.Error("workout create encode", "err", err)
		}
	}
}

// GetWorkout returns a workout by id.
func (a *API) GetWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		workout, err := a.Workouts.Get(r.Context(), id)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout get", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, workout); err != nil {
			a.Logger.Error("workout get encode", "err", err)
		}
	}
}

// ExportWorkout returns a workout with nested steps/exercises for sharing.
func (a *API) ExportWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		workout, err := a.Workouts.Export(r.Context(), id)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout export", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, workout); err != nil {
			a.Logger.Error("workout export encode", "err", err)
		}
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
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout import decode", "err", err)
			}
			return
		}

		resolvedUserID, err := a.resolveUserID(r, req.UserID)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout import user", "err", err)
			}
			return
		}
		req.UserID = resolvedUserID

		created, err := a.Workouts.Import(r.Context(), req.UserID, req.Workout)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout import", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusCreated, created); err != nil {
			a.Logger.Error("workout import encode", "err", err)
		}
	}
}

// UpdateWorkout replaces a workout and its steps.
func (a *API) UpdateWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		req, err := decode[workouts.WorkoutRequest](r)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout update decode", "err", err)
			}
			return
		}

		if a.AuthHeader != "" {
			resolvedUserID, err := a.resolveUserID(r, "")
			if err != nil {
				if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
					a.Logger.Error("workout update user", "err", err)
				}
				return
			}
			req.UserID = resolvedUserID
		} else if req.UserID == "" {
			current, err := a.Workouts.Get(r.Context(), id)
			if err != nil {
				if err := encode(w, r, http.StatusNotFound, apiError{Error: err.Error()}); err != nil {
					a.Logger.Error("workout update lookup", "err", err)
				}
				return
			}
			req.UserID = current.UserID
		}

		updated, err := a.Workouts.Update(r.Context(), id, req)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout update", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, updated); err != nil {
			a.Logger.Error("workout update encode", "err", err)
		}
	}
}

// DeleteWorkout removes a workout by id.
func (a *API) DeleteWorkout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		if err := a.Workouts.Delete(r.Context(), id); err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("workout delete", "err", err)
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
