package handler

import (
	"net/http"
	"time"

	"github.com/gi8lino/motus/internal/service/sessions"
)

// CreateSession initializes a new session state from a workout.
func (a *API) CreateSession() http.HandlerFunc {
	type createSessionRequest struct {
		WorkoutID string `json:"workoutId"`
	}
	type createSessionResponse struct {
		SessionID string                `json:"sessionId"`
		State     sessions.SessionState `json:"state"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[createSessionRequest](r)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("session create decode", "err", err)
			}
			return
		}

		state, err := a.Sessions.CreateState(r.Context(), req.WorkoutID)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("session create", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusCreated, createSessionResponse{
			SessionID: state.SessionID,
			State:     state,
		}); err != nil {
			a.Logger.Error("session create encode", "err", err)
		}
	}
}

// ListSessionHistory returns completed sessions for the current user.
func (a *API) ListSessionHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		resolvedID, err := a.resolveUserID(r, userID)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("session history user", "err", err)
			}
			return
		}

		items, err := a.Sessions.BuildSessionHistory(r.Context(), resolvedID, 25)
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("session history", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, items); err != nil {
			a.Logger.Error("session history encode", "err", err)
		}
	}
}

// SessionSteps returns stored step timings for a session.
func (a *API) SessionSteps() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		steps, err := a.Sessions.FetchStepTimings(r.Context(), r.PathValue("id"))
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("session steps", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusOK, steps); err != nil {
			a.Logger.Error("session steps encode", "err", err)
		}
	}
}

// CompleteSession records a completed session and its step timings.
func (a *API) CompleteSession() http.HandlerFunc {
	type completeSessionRequest struct {
		SessionID   string                      `json:"sessionId"`
		WorkoutID   string                      `json:"workoutId"`
		WorkoutName string                      `json:"workoutName"`
		UserID      string                      `json:"userId"`
		StartedAt   time.Time                   `json:"startedAt"`
		CompletedAt time.Time                   `json:"completedAt"`
		Steps       []sessions.SessionStepState `json:"steps"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decode[completeSessionRequest](r)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("session complete decode", "err", err)
			}
			return
		}

		resolvedUserID, err := a.resolveUserID(r, req.UserID)
		if err != nil {
			if err := encode(w, r, http.StatusBadRequest, apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("session complete user", "err", err)
			}
			return
		}
		req.UserID = resolvedUserID

		log, err := a.Sessions.RecordSession(r.Context(), sessions.CompleteRequest{
			SessionID:   req.SessionID,
			WorkoutID:   req.WorkoutID,
			WorkoutName: req.WorkoutName,
			UserID:      req.UserID,
			StartedAt:   req.StartedAt,
			CompletedAt: req.CompletedAt,
			Steps:       req.Steps,
		})
		if err != nil {
			if err := encode(w, r, serviceStatus(err), apiError{Error: err.Error()}); err != nil {
				a.Logger.Error("session complete", "err", err)
			}
			return
		}

		if err := encode(w, r, http.StatusCreated, log); err != nil {
			a.Logger.Error("session complete encode", "err", err)
		}
	}
}
