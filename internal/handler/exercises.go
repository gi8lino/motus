package handler

import "net/http"

// ListExercises returns the exercise catalog for the current user.
func (a *API) ListExercises() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.resolveUserID(r, "")
		if err != nil {
			a.logRequestError(r, "resolve_user_id_failed", "resolve user id failed", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		items, err := a.Exercises.List(r.Context(), userID)
		if err != nil {
			a.logRequestError(r, "list_exercises_failed", "list exercises failed", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.respondJSON(w, http.StatusOK, items)
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
			a.logRequestError(r, "resolve_user_id_failed", "resolve user id failed", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		req, err := decode[createExerciseRequest](r)
		if err != nil {
			a.logRequestError(r, "decode_request_failed", "decode request failed", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		exercise, err := a.Exercises.Create(r.Context(), userID, req.Name, req.IsCore)
		if err != nil {
			a.logRequestError(r, "create_exercise_failed", "create exercise failed", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.businessLogger(r).Info("exercise created",
			"event", "exercise_created",
			"resource", "exercise",
			"resource_id", exercise.ID,
			"user_id", userID,
			"is_core", exercise.IsCore,
		)
		a.respondJSON(w, http.StatusCreated, exercise)
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
			a.logRequestError(r, "resolve_user_id_failed", "resolve user id failed", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		req, err := decode[updateExerciseRequest](r)
		if err != nil {
			a.logRequestError(r, "decode_request_failed", "decode request failed", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		updated, err := a.Exercises.Update(r.Context(), userID, id, req.Name)
		if err != nil {
			a.logRequestError(r, "update_exercise_failed", "update exercise failed", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.businessLogger(r).Info("exercise updated",
			"event", "exercise_updated",
			"resource", "exercise",
			"resource_id", updated.ID,
			"user_id", userID,
		)
		a.respondJSON(w, http.StatusOK, updated)
	}
}

// DeleteExercise removes an exercise from the catalog.
func (a *API) DeleteExercise() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		userID, err := a.resolveUserID(r, "")
		if err != nil {
			a.logRequestError(r, "resolve_user_id_failed", "exercise delete user", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		if err := a.Exercises.Delete(r.Context(), userID, id); err != nil {
			a.logRequestError(r, "delete_exercise_failed", "delete exercise failed", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.businessLogger(r).Info("exercise deleted",
			"event", "exercise_deleted",
			"resource", "exercise",
			"resource_id", id,
			"user_id", userID,
		)
		a.respondJSON(w, http.StatusNoContent, statusResponse{Status: "ok"})
	}
}

// BackfillExercises rebuilds core exercises from workout data.
func (a *API) BackfillExercises() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := a.Exercises.Backfill(r.Context()); err != nil {
			a.logRequestError(r, "backfill_exercises_failed", "backfill exercises failed", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.businessLogger(r).Info("exercise backfill complete",
			"event", "exercise_backfill",
			"resource", "exercise",
		)
		a.respondJSON(w, http.StatusOK, statusResponse{Status: "ok"})
	}
}
