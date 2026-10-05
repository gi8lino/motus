package endpoint

import (
	"github.com/gi8lino/motus/internal/http/response"
	"log/slog"
	"net/http"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service/workouts"
)

// GetWorkouts lists workouts for the current user.
func GetWorkouts(service *workouts.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		resolvedID, err := currentUser(r, userID)
		if err != nil {
			logger.Error("resolve user id failed", "event", "resolve_user_id_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		workouts, err := service.List(r.Context(), resolvedID)
		if err != nil {
			logger.Error("list workouts failed", "event", "list_workouts_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		response.Respond(w, http.StatusOK, workouts)
	}
}

// CreateWorkout stores a new workout for the current user.
func CreateWorkout(service *workouts.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		req, err := decode[workouts.WorkoutRequest](r)
		if err != nil {
			logger.Error("decode request failed", "event", "decode_request_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		resolvedUserID, err := currentUser(r, userID)
		if err != nil {
			logger.Error("resolve user id failed", "event", "resolve_user_id_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}
		req.UserID = resolvedUserID

		created, err := service.Create(r.Context(), req)
		if err != nil {
			logger.Error("create workout failed", "event", "create_workout_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("workout created",
			"event", "workout_created",
			"resource", "workout",
			"resource_id", created.ID,
			"user_id", created.UserID,
			"count", len(created.Steps),
		)
		response.Respond(w, http.StatusCreated, created)
	}
}

// GetWorkout returns a workout by id.
func GetWorkout(service *workouts.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := requireWorkoutOwner(service, r, id); err != nil {
			response.Problem(w, http.StatusNotFound, "workout not found")
			return
		}

		workout, err := service.Get(r.Context(), id)
		if err != nil {
			logger.Error("get workout failed", "event", "get_workout_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		response.Respond(w, http.StatusOK, workout)
	}
}

// ExportWorkout returns a workout with nested steps/exercises for sharing.
func ExportWorkout(service *workouts.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := requireWorkoutOwner(service, r, id); err != nil {
			response.Problem(w, http.StatusNotFound, "workout not found")
			return
		}

		workout, err := service.Export(r.Context(), id)
		if err != nil {
			logger.Error("export workout failed", "event", "export_workout_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("workout exported",
			"event", "workout_exported",
			"resource", "workout",
			"resource_id", workout.ID,
			"user_id", workout.UserID,
		)
		response.Respond(w, http.StatusOK, workout)
	}
}

// ImportWorkout creates a new workout from exported JSON.
func ImportWorkout(service *workouts.Service, logger *slog.Logger) http.HandlerFunc {
	type importWorkoutRequest struct {
		UserID  string     `json:"userId"`
		Workout db.Workout `json:"workout"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[importWorkoutRequest](r)
		if err != nil {
			logger.Error("decode request failed", "event", "decode_request_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		resolvedUserID, err := currentUser(r, req.UserID)
		if err != nil {
			logger.Error("resolve user id failed", "event", "resolve_user_id_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}
		req.UserID = resolvedUserID

		created, err := service.Import(r.Context(), req.UserID, req.Workout)
		if err != nil {
			logger.Error("import workout failed", "event", "import_workout_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("workout imported",
			"event", "workout_imported",
			"resource", "workout",
			"resource_id", created.ID,
			"user_id", created.UserID,
		)
		response.Respond(w, http.StatusCreated, created)
	}
}

// UpdateWorkout replaces a workout and its steps.
func UpdateWorkout(service *workouts.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := requireWorkoutOwner(service, r, id); err != nil {
			response.Problem(w, http.StatusNotFound, "workout not found")
			return
		}

		req, err := decode[workouts.WorkoutRequest](r)
		if err != nil {
			logger.Error("decode request failed", "event", "decode_request_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		if true {
			resolvedUserID, err := currentUser(r, "")
			if err != nil {
				logger.Error("resolve user id failed", "event", "resolve_user_id_failed", "err", err)
				response.Problem(w, http.StatusBadRequest, err.Error())
				return
			}
			req.UserID = resolvedUserID
		} else if req.UserID == "" {
			current, err := service.Get(r.Context(), id)
			if err != nil {
				logger.Error("get workout failed", "event", "get_workout_failed", "err", err)
				response.Problem(w, http.StatusNotFound, err.Error())
				return
			}
			req.UserID = current.UserID
		}

		updated, err := service.Update(r.Context(), id, req)
		if err != nil {
			logger.Error("update workout failed", "event", "update_workout_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("workout updated",
			"event", "workout_updated",
			"resource", "workout",
			"resource_id", updated.ID,
			"user_id", updated.UserID,
			"count", len(updated.Steps),
		)
		response.Respond(w, http.StatusOK, updated)
	}
}

// DeleteWorkout removes a workout by id.
func DeleteWorkout(service *workouts.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := requireWorkoutOwner(service, r, id); err != nil {
			response.Problem(w, http.StatusNotFound, "workout not found")
			return
		}

		userID := "test@example.com"
		if true {
			var err error
			userID, err = currentUser(r, "")
			if err != nil {
				response.Problem(w, http.StatusUnauthorized, "authentication required")
				return
			}
		}
		if err := service.Delete(r.Context(), id, userID); err != nil {
			logger.Error("delete workout failed", "event", "delete_workout_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("workout deleted",
			"event", "workout_deleted",
			"resource", "workout",
			"resource_id", id,
		)
		response.Respond(w, http.StatusNoContent, statusResponse{Status: "ok"})
	}
}
