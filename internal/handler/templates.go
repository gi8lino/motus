package handler

import (
	"net/http"

	"github.com/gi8lino/motus/internal/service/templates"
)

// ListTemplates returns all shared templates.
func (a *API) ListTemplates() http.HandlerFunc {
	svc := templates.New(a.Store)
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := svc.List(r.Context())
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// CreateTemplate marks a workout as a template.
func (a *API) CreateTemplate() http.HandlerFunc {
	svc := templates.New(a.Store)
	type createTemplateRequest struct {
		WorkoutID string `json:"workoutId"`
		Name      string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[createTemplateRequest](r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		template, err := svc.Create(r.Context(), req.WorkoutID, req.Name)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, template)
	}
}

// GetTemplate returns a template by id.
func (a *API) GetTemplate() http.HandlerFunc {
	svc := templates.New(a.Store)
	return func(w http.ResponseWriter, r *http.Request) {
		template, err := svc.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, template)
	}
}

// ApplyTemplate clones a template into a new workout.
func (a *API) ApplyTemplate() http.HandlerFunc {
	svc := templates.New(a.Store)
	type applyTemplateRequest struct {
		UserID string `json:"userId"`
		Name   string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[applyTemplateRequest](r)
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

		workout, err := svc.Apply(r.Context(), r.PathValue("id"), req.UserID, req.Name)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, workout)
	}
}
