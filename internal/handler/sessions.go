package handler

import (
	"net/http"
	"time"

	"github.com/gi8lino/motus/internal/service/sessions"
	"github.com/gi8lino/motus/internal/service/sounds"
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
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		state, err := sessions.CreateState(r.Context(), a.Store, req.WorkoutID, sounds.URLByKey)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, createSessionResponse{
			SessionID: state.SessionID,
			State:     state,
		})
	}
}

// ListSessionHistory returns completed sessions for the current user.
func (a *API) ListSessionHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")

		resolvedID, err := a.resolveUserID(r, userID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		history, err := a.Store.SessionHistory(r.Context(), resolvedID, 25)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}

		items, err := sessions.BuildSessionHistory(r.Context(), a.Store, history)
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, items)
	}
}

// SessionSteps returns stored step timings for a session.
func (a *API) SessionSteps() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		steps, err := sessions.FetchStepTimings(r.Context(), a.Store, r.PathValue("id"))
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, steps)
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
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}

		resolvedUserID, err := a.resolveUserID(r, req.UserID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.UserID = resolvedUserID

		log, err := sessions.RecordSession(r.Context(), a.Store, sessions.CompleteRequest{
			SessionID:   req.SessionID,
			WorkoutID:   req.WorkoutID,
			WorkoutName: req.WorkoutName,
			UserID:      req.UserID,
			StartedAt:   req.StartedAt,
			CompletedAt: req.CompletedAt,
			Steps:       req.Steps,
		})
		if err != nil {
			writeJSON(w, serviceStatus(err), apiError{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusCreated, log)
	}
}
