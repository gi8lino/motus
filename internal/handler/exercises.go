package handler

import (
	"net/http"

	"github.com/gi8lino/motus/internal/service/exercises"
)

// ListExercises returns the exercise catalog for the current user.
func (a *API) ListExercises() http.HandlerFunc {
	svc := exercises.New(a.ExercisesStore)
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		items, err := svc.List(r.Context(), userID)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateExercise adds a new exercise to the catalog.
func (a *API) CreateExercise() http.HandlerFunc {
	svc := exercises.New(a.ExercisesStore)
	type createExerciseRequest struct {
		Name   string `json:"name"`
		IsCore bool   `json:"isCore"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		req, err := decode[createExerciseRequest](r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		exercise, err := svc.Create(r.Context(), userID, req.Name, req.IsCore)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, exercise)
	}
}

// UpdateExercise renames an exercise or creates a personal copy.
func (a *API) UpdateExercise() http.HandlerFunc {
	svc := exercises.New(a.ExercisesStore)
	type updateExerciseRequest struct {
		Name string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		userID, err := a.resolveUserID(r, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		req, err := decode[updateExerciseRequest](r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		updated, err := svc.Update(r.Context(), userID, id, req.Name)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, updated)
	}
}

// DeleteExercise removes an exercise from the catalog.
func (a *API) DeleteExercise() http.HandlerFunc {
	svc := exercises.New(a.ExercisesStore)
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		userID, err := a.resolveUserID(r, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		if err := svc.Delete(r.Context(), userID, id); err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// BackfillExercises rebuilds core exercises from workout data.
func (a *API) BackfillExercises() http.HandlerFunc {
	svc := exercises.New(a.ExercisesStore)
	return func(w http.ResponseWriter, r *http.Request) {
		if err := svc.Backfill(r.Context()); err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	}
}
