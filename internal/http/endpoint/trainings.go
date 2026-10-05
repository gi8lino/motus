package endpoint

import (
	"github.com/gi8lino/motus/internal/http/response"
	"log/slog"
	"net/http"

	"github.com/gi8lino/motus/internal/service/workouts"
	"time"

	"github.com/gi8lino/motus/internal/service/trainings"
)

// CreateTraining initializes a new training state from a workout.
func CreateTraining(service *trainings.Service, workoutService *workouts.Service, logger *slog.Logger) http.HandlerFunc {
	type createTrainingRequest struct {
		WorkoutID string `json:"workoutId"`
	}
	type createTrainingResponse struct {
		TrainingID string                  `json:"trainingId"`
		State      trainings.TrainingState `json:"state"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[createTrainingRequest](r)
		if err != nil {
			logger.Error("decode request failed", "event", "decode_request_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := requireWorkoutOwner(workoutService, r, req.WorkoutID); err != nil {
			response.Problem(w, http.StatusNotFound, "workout not found")
			return
		}

		state, err := service.CreateState(r.Context(), req.WorkoutID)
		if err != nil {
			logger.Error("create training state failed", "event", "create_training_state_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("training created",
			"event", "training_created",
			"resource", "training",
			"resource_id", state.TrainingID,
			"user_id", state.UserID,
			"workout_id", state.WorkoutID,
		)
		response.Respond(w, http.StatusCreated, createTrainingResponse{
			TrainingID: state.TrainingID,
			State:      state,
		})
	}
}

// ListTrainingHistory returns completed trainings for the current user.
func ListTrainingHistory(service *trainings.Service, workoutService *workouts.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		resolvedID, err := currentUser(r, userID)
		if err != nil {
			logger.Error("resolve user id failed", "event", "resolve_user_id_failed", "err", err)
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}

		items, err := service.BuildTrainingHistory(r.Context(), resolvedID, 25)
		if err != nil {
			logger.Error("build training history failed", "event", "build_training_history_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		response.Respond(w, http.StatusOK, items)
	}
}

// TrainingSteps returns stored step timings for a training.
func TrainingSteps(service *trainings.Service, workoutService *workouts.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		steps, err := service.FetchStepTimings(r.Context(), r.PathValue("id"))
		if err != nil {
			logger.Error("fetch step timings failed", "event", "fetch_step_timings_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		response.Respond(w, http.StatusOK, steps)
	}
}

// CompleteTraining records a completed training and its step timings.
func CompleteTraining(service *trainings.Service, workoutService *workouts.Service, logger *slog.Logger) http.HandlerFunc {
	type completeTrainingRequest struct {
		TrainingID  string                        `json:"trainingId"`
		WorkoutID   string                        `json:"workoutId"`
		WorkoutName string                        `json:"workoutName"`
		UserID      string                        `json:"userId"`
		StartedAt   time.Time                     `json:"startedAt"`
		CompletedAt time.Time                     `json:"completedAt"`
		Steps       []trainings.TrainingStepState `json:"steps"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[completeTrainingRequest](r)
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
		if err := requireWorkoutOwner(workoutService, r, req.WorkoutID); err != nil {
			response.Problem(w, http.StatusNotFound, "workout not found")
			return
		}

		log, err := service.RecordTraining(r.Context(), trainings.CompleteRequest{
			TrainingID:  req.TrainingID,
			WorkoutID:   req.WorkoutID,
			WorkoutName: req.WorkoutName,
			UserID:      req.UserID,
			StartedAt:   req.StartedAt,
			CompletedAt: req.CompletedAt,
			Steps:       req.Steps,
		})
		if err != nil {
			logger.Error("record training failed", "event", "record_training_failed", "err", err)
			response.ServiceError(logger, w, err)
			return
		}

		logger.Info("training completed",
			"event", "training_completed",
			"resource", "training",
			"resource_id", log.ID,
			"user_id", log.UserID,
			"workout_id", log.WorkoutID,
			"count", len(req.Steps),
		)
		response.Respond(w, http.StatusCreated, log)
	}
}

// UpdateTrainingFeedback stores notes and perceived effort for a completed training.
func UpdateTrainingFeedback(service *trainings.Service, workoutService *workouts.Service, logger *slog.Logger) http.HandlerFunc {
	type request struct {
		UserID          string `json:"userId"`
		Notes           string `json:"notes"`
		PerceivedEffort *int   `json:"perceivedEffort"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[request](r)
		if err != nil {
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}
		userID, err := currentUser(r, req.UserID)
		if err != nil {
			response.Problem(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := service.UpdateFeedback(r.Context(), r.PathValue("id"), userID, req.Notes, req.PerceivedEffort); err != nil {
			response.ServiceError(logger, w, err)
			return
		}
		response.Respond(w, http.StatusNoContent, nil)
	}
}
