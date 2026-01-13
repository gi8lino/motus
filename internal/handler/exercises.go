package handler

import "net/http"

// ListExercises returns the exercise catalog for the current user.
func (a *API) ListExercises() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("exercise list user", "err", err)
			}
			return
		}

		items, err := a.Exercises.List(r.Context(), userID)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("exercise list", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, items); err != nil {
			a.Logger.Error("exercise list encode", "err", err)
		}
	}
}

// CreateExercise adds a new exercise to the catalog.
func (a *API) CreateExercise() http.HandlerFunc {
	type createExerciseRequest struct {
		Name   string `json:"name"`
		IsCore bool   `json:"isCore"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("exercise create user", "err", err)
			}
			return
		}

		req, err := decode[createExerciseRequest](r)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("exercise create decode", "err", err)
			}
			return
		}

		exercise, err := a.Exercises.Create(r.Context(), userID, req.Name, req.IsCore)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("exercise create", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusCreated, exercise); err != nil {
			a.Logger.Error("exercise create encode", "err", err)
		}
	}
}

// UpdateExercise renames an exercise or creates a personal copy.
func (a *API) UpdateExercise() http.HandlerFunc {
	type updateExerciseRequest struct {
		Name string `json:"name"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		userID, err := a.resolveUserID(r, "")
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("exercise update user", "err", err)
			}
			return
		}

		req, err := decode[updateExerciseRequest](r)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("exercise update decode", "err", err)
			}
			return
		}

		updated, err := a.Exercises.Update(r.Context(), userID, id, req.Name)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("exercise update", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, updated); err != nil {
			a.Logger.Error("exercise update encode", "err", err)
		}
	}
}

// DeleteExercise removes an exercise from the catalog.
func (a *API) DeleteExercise() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		userID, err := a.resolveUserID(r, "")
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("exercise delete user", "err", err)
			}
			return
		}

		if err := a.Exercises.Delete(r.Context(), userID, id); err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("exercise delete", "err", err)
			}
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// BackfillExercises rebuilds core exercises from workout data.
func (a *API) BackfillExercises() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := a.Exercises.Backfill(r.Context()); err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("exercise backfill", "err", err)
			}
			return
		}
		if err := encode(w, r, http.StatusOK, statusResponse{Status: "ok"}); err != nil {
			a.Logger.Error("exercise backfill encode", "err", err)
		}
	}
}
