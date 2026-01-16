package handler

import (
	"net/http"
	"time"

	"github.com/gi8lino/motus/internal/service/trainings"
)

// CreateTraining initializes a new training state from a workout.
func (a *API) CreateTraining() http.HandlerFunc {
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
			a.Logger.Error("decode request failed", "err", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		state, err := a.Trainings.CreateState(r.Context(), req.WorkoutID)
		if err != nil {
			a.Logger.Error("create training state failed", "err", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.Logger.Debug("create training state", "state", state)
		a.respondJSON(w, http.StatusCreated, createTrainingResponse{
			TrainingID: state.TrainingID,
			State:      state,
		})
	}
}

// ListTrainingHistory returns completed trainings for the current user.
func (a *API) ListTrainingHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		resolvedID, err := a.resolveUserID(r, userID)
		if err != nil {
			a.Logger.Error("resolve user id failed", "err", err)
			a.respondJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		items, err := a.Trainings.BuildTrainingHistory(r.Context(), resolvedID, 25)
		if err != nil {
			a.Logger.Error("build training history failed", "err", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.Logger.Error("build training history", "items", items)
		a.respondJSON(w, http.StatusOK, items)
	}
}

// TrainingSteps returns stored step timings for a training.
func (a *API) TrainingSteps() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		steps, err := a.Trainings.FetchStepTimings(r.Context(), r.PathValue("id"))
		if err != nil {
			a.Logger.Error("fetch step timings failed", "err", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.Logger.Error("fetch step timings", "steps", steps)
		a.respondJSON(w, http.StatusOK, steps)
	}
}

// CompleteTraining records a completed training and its step timings.
func (a *API) CompleteTraining() http.HandlerFunc {
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

		log, err := a.Trainings.RecordTraining(r.Context(), trainings.CompleteRequest{
			TrainingID:  req.TrainingID,
			WorkoutID:   req.WorkoutID,
			WorkoutName: req.WorkoutName,
			UserID:      req.UserID,
			StartedAt:   req.StartedAt,
			CompletedAt: req.CompletedAt,
			Steps:       req.Steps,
		})
		if err != nil {
			a.Logger.Error("record training failed", "err", err)
			a.respondJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		a.Logger.Debug("record training", "log", log)
		a.respondJSON(w, http.StatusCreated, log)
	}
}
