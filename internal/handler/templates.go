package handler

import "net/http"

// ListTemplates returns all shared templates.
func (a *API) ListTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := a.Templates.List(r.Context())
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("template list", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, items); err != nil {
			a.Logger.Error("template list encode", "err", err)
		}
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
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("template create decode", "err", err)
			}
			return
		}

		template, err := a.Templates.Create(r.Context(), req.WorkoutID, req.Name)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("template create", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusCreated, template); err != nil {
			a.Logger.Error("template create encode", "err", err)
		}
	}
}

// GetTemplate returns a template by id.
func (a *API) GetTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		template, err := a.Templates.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("template get", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, template); err != nil {
			a.Logger.Error("template get encode", "err", err)
		}
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
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("template apply decode", "err", err)
			}
			return
		}

		resolvedUserID, err := a.resolveUserID(r, req.UserID)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("template apply user", "err", err)
			}
			return
		}
		req.UserID = resolvedUserID

		workout, err := a.Templates.Apply(r.Context(), r.PathValue("id"), req.UserID, req.Name)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("template apply", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusCreated, workout); err != nil {
			a.Logger.Error("template apply encode", "err", err)
		}
	}
}
