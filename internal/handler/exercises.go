package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ListExercises returns the exercise catalog for the current user.
func (a *API) ListExercises() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		exercises, err := a.Store.ListExercises(r.Context(), userID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, exercises)
	}
}

// CreateExercise adds a new exercise to the catalog.
func (a *API) CreateExercise() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		user, err := a.Store.GetUser(r.Context(), userID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, apiError{Error: "user not found"})
			return
		}
		var req struct {
			Name   string `json:"name"`
			IsCore bool   `json:"isCore"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		if req.IsCore && !user.IsAdmin {
			writeJSON(w, http.StatusForbidden, apiError{Error: "core exercise requires admin privileges"})
			return
		}
		exercise, err := a.Store.CreateExercise(r.Context(), req.Name, userID, req.IsCore)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, exercise)
	}
}

// UpdateExercise renames an exercise or creates a personal copy.
func (a *API) UpdateExercise() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "exercise id is required"})
			return
		}
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		user, err := a.Store.GetUser(r.Context(), userID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, apiError{Error: "user not found"})
			return
		}
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		exercise, err := a.Store.GetExercise(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, apiError{Error: "exercise not found"})
			return
		}
		if exercise.IsCore && !user.IsAdmin {
			copied, err := a.Store.CreateExercise(r.Context(), req.Name, userID, false)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
				return
			}
			if err := a.Store.ReplaceExerciseForUser(r.Context(), userID, exercise.ID, copied.ID, copied.Name); err != nil {
				writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, copied)
			return
		}
		if !user.IsAdmin && exercise.OwnerUserID != "" && exercise.OwnerUserID != userID {
			writeJSON(w, http.StatusForbidden, apiError{Error: "exercise not owned by user"})
			return
		}
		updated, err := a.Store.RenameExercise(r.Context(), id, req.Name)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

// DeleteExercise removes an exercise from the catalog.
func (a *API) DeleteExercise() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "exercise id is required"})
			return
		}
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		user, err := a.Store.GetUser(r.Context(), userID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, apiError{Error: "user not found"})
			return
		}
		exercise, err := a.Store.GetExercise(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, apiError{Error: "exercise not found"})
			return
		}
		if exercise.IsCore && !user.IsAdmin {
			writeJSON(w, http.StatusForbidden, apiError{Error: "cannot delete core exercise"})
			return
		}
		if !user.IsAdmin && exercise.OwnerUserID != "" && exercise.OwnerUserID != userID {
			writeJSON(w, http.StatusForbidden, apiError{Error: "exercise not owned by user"})
			return
		}
		if err := a.Store.DeleteExercise(r.Context(), id); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// BackfillExercises rebuilds core exercises from workout data.
func (a *API) BackfillExercises() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := a.Store.BackfillCoreExercises(r.Context()); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
