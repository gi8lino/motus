package handler

import "net/http"

// ListTemplates returns all shared templates.
func (a *API) ListTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := a.Templates.List(r.Context())
		if err != nil {
			a.Logger.Error("list templates failed", "err", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.respondJSON(w, http.StatusOK, items)
	}
}

// CreateTemplate marks a workout as a template.
func (a *API) CreateTemplate() http.HandlerFunc {
	type createTemplateRequest struct {
		WorkoutID string `json:"workoutId"`
		Name      string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[createTemplateRequest](r)
		if err != nil {
			a.Logger.Error("decode request failed", "err", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		template, err := a.Templates.Create(r.Context(), req.WorkoutID, req.Name)
		if err != nil {
			a.Logger.Error("create template failed", "err", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.respondJSON(w, http.StatusCreated, template)
	}
}

// GetTemplate returns a template by id.
func (a *API) GetTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		template, err := a.Templates.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			a.Logger.Error("get template failed", "err", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.respondJSON(w, http.StatusOK, template)
	}
}

// ApplyTemplate clones a template into a new workout.
func (a *API) ApplyTemplate() http.HandlerFunc {
	type applyTemplateRequest struct {
		UserID string `json:"userId"`
		Name   string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[applyTemplateRequest](r)
		if err != nil {
			a.Logger.Error("decode request failed", "err", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		resolvedUserID, err := a.resolveUserID(r, req.UserID)
		if err != nil {
			a.Logger.Error("resolve user id failed", "err", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.UserID = resolvedUserID

		workout, err := a.Templates.Apply(r.Context(), r.PathValue("id"), req.UserID, req.Name)
		if err != nil {
			a.Logger.Error("apply template failed", "err", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.Logger.Debug("apply template", "workout", workout)
		a.respondJSON(w, http.StatusCreated, workout)
	}
}
