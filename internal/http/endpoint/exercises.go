package endpoint

import (
	"github.com/gi8lino/motus/internal/http/response"
	"log/slog"
	"net/http"

	"github.com/gi8lino/motus/internal/service/exercises"
)

// ListExercises returns the exercise catalog for the current user.
func ListExercises(service *exercises.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := currentUser(r, "")
		if err != nil {
			logger.Error("resolve user id failed", "event", "resolve_user_id_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		items, err := service.List(r.Context(), userID)
		if err != nil {
			logger.Error("list exercises failed", "event", "list_exercises_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		response.Respond(w, http.StatusOK, items)
	}
}

// CreateExercise adds a new exercise to the catalog.
func CreateExercise(service *exercises.Service, logger *slog.Logger) http.HandlerFunc {
	type createExerciseRequest struct {
		Name   string `json:"name"`
		IsCore bool   `json:"isCore"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := currentUser(r, "")
		if err != nil {
			logger.Error("resolve user id failed", "event", "resolve_user_id_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		req, err := decode[createExerciseRequest](r)
		if err != nil {
			logger.Error("decode request failed", "event", "decode_request_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		exercise, err := service.Create(r.Context(), userID, req.Name, req.IsCore)
		if err != nil {
			logger.Error("create exercise failed", "event", "create_exercise_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("exercise created",
			"event", "exercise_created",
			"resource", "exercise",
			"resource_id", exercise.ID,
			"user_id", userID,
			"is_core", exercise.IsCore,
		)
		response.Respond(w, http.StatusCreated, exercise)
	}
}

// UpdateExercise renames an exercise or creates a personal copy.
func UpdateExercise(service *exercises.Service, logger *slog.Logger) http.HandlerFunc {
	type updateExerciseRequest struct {
		Name     string `json:"name"`
		HasSides *bool  `json:"hasSides"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		userID, err := currentUser(r, "")
		if err != nil {
			logger.Error("resolve user id failed", "event", "resolve_user_id_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		req, err := decode[updateExerciseRequest](r)
		if err != nil {
			logger.Error("decode request failed", "event", "decode_request_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		updated, err := service.Update(r.Context(), userID, id, req.Name)
		if err != nil {
			logger.Error("update exercise failed", "event", "update_exercise_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}
		if req.HasSides != nil {
			updated, err = service.SetHasSides(r.Context(), userID, id, *req.HasSides)
			if err != nil {
				response.ServiceError(logger, w, err)
				return
			}
		}

		logger.Info("exercise updated",
			"event", "exercise_updated",
			"resource", "exercise",
			"resource_id", updated.ID,
			"user_id", userID,
		)
		response.Respond(w, http.StatusOK, updated)
	}
}

// DeleteExercise removes an exercise from the catalog.
func DeleteExercise(service *exercises.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		userID, err := currentUser(r, "")
		if err != nil {
			logger.Error("exercise delete user", "event", "resolve_user_id_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := service.Delete(r.Context(), userID, id); err != nil {
			logger.Error("delete exercise failed", "event", "delete_exercise_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("exercise deleted",
			"event", "exercise_deleted",
			"resource", "exercise",
			"resource_id", id,
			"user_id", userID,
		)
		response.Respond(w, http.StatusNoContent, statusResponse{Status: "ok"})
	}
}

// BackfillExercises rebuilds core exercises from workout data.
func BackfillExercises(service *exercises.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := service.Backfill(r.Context()); err != nil {
			logger.Error("backfill exercises failed", "event", "backfill_exercises_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("exercise backfill complete",
			"event", "exercise_backfill",
			"resource", "exercise",
		)
		response.Respond(w, http.StatusOK, statusResponse{Status: "ok"})
	}
}
