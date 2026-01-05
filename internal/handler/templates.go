package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ListTemplates returns all shared templates.
func (a *API) ListTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		templates, err := a.Store.ListTemplates(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, templates)
	}
}

// CreateTemplate marks a workout as a template.
func (a *API) CreateTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			WorkoutID string `json:"workoutId"`
			Name      string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.WorkoutID = strings.TrimSpace(req.WorkoutID)
		if req.WorkoutID == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "workoutId is required"})
			return
		}
		template, err := a.Store.CreateTemplateFromWorkout(r.Context(), req.WorkoutID, req.Name)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, template)
	}
}

// GetTemplate returns a template by id.
func (a *API) GetTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "template id is required"})
			return
		}
		template, err := a.Store.WorkoutWithSteps(r.Context(), id)
		if err != nil || !template.IsTemplate {
			if err == nil {
				writeJSON(w, http.StatusNotFound, apiError{Error: "template not found"})
				return
			}
			writeJSON(w, http.StatusNotFound, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, template)
	}
}

// ApplyTemplate clones a template into a new workout.
func (a *API) ApplyTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSpace(r.PathValue("id"))
		if id == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "template id is required"})
			return
		}
		var req struct {
			UserID string `json:"userId"`
			Name   string `json:"name"`
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
		workout, err := a.Store.CreateWorkoutFromTemplate(r.Context(), id, req.UserID, req.Name)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, workout)
	}
}
