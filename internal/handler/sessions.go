package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

type sessionHistoryItem struct {
	ID          string              `json:"id"`
	SessionID   string              `json:"sessionId"`
	WorkoutID   string              `json:"workoutId"`
	WorkoutName string              `json:"workoutName"`
	UserID      string              `json:"userId"`
	StartedAt   *time.Time          `json:"startedAt,omitempty"`
	CompletedAt *time.Time          `json:"completedAt,omitempty"`
	Steps       []db.SessionStepLog `json:"steps,omitempty"`
}

// Session types mirror the JSON contract with the SPA.
type SessionState struct {
	SessionID    string             `json:"sessionId"`
	WorkoutID    string             `json:"workoutId"`
	UserID       string             `json:"userId"`
	WorkoutName  string             `json:"workoutName"`
	CurrentIndex int                `json:"currentIndex"`
	Running      bool               `json:"running"`
	Done         bool               `json:"done"`
	StartedAt    time.Time          `json:"startedAt"`
	CompletedAt  time.Time          `json:"completedAt"`
	Steps        []SessionStepState `json:"steps"`
}

type SessionStepState struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	Type             string       `json:"type"`
	EstimatedSeconds int          `json:"estimatedSeconds"`
	SoundURL         string       `json:"soundUrl"`
	Running          bool         `json:"running"`
	Completed        bool         `json:"completed"`
	Current          bool         `json:"current"`
	ElapsedMillis    int64        `json:"elapsedMillis"`
	Exercises        []Exercise   `json:"exercises"`
	PauseOptions     PauseOptions `json:"pauseOptions"`
}

type Exercise struct {
	Name   string `json:"name"`
	Amount string `json:"amount"`
	Weight string `json:"weight"`
}

type PauseOptions struct {
	AutoAdvance bool `json:"autoAdvance"`
}

// CreateSession initializes a new session state from a workout.
func (a *API) CreateSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			WorkoutID string `json:"workoutId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.WorkoutID = strings.TrimSpace(req.WorkoutID)
		if req.WorkoutID == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "workoutId is required"})
			return
		}
		workout, err := a.Store.WorkoutWithSteps(r.Context(), req.WorkoutID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, apiError{Error: err.Error()})
			return
		}
		state := sessionStateFromWorkout(workout)
		sessionID := state.SessionID
		writeJSON(w, http.StatusCreated, map[string]any{
			"sessionId": sessionID,
			"state":     state,
		})
	}
}

// ListSessionHistory returns completed sessions for the current user.
func (a *API) ListSessionHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := strings.TrimSpace(r.PathValue("id"))
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
		stepMap := make(map[string][]db.SessionStepLog, len(history))
		for _, entry := range history {
			steps, err := a.Store.SessionStepTimings(r.Context(), entry.ID)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
				return
			}
			stepMap[entry.ID] = steps
		}
		items := buildSessionHistoryItems(history, stepMap)
		writeJSON(w, http.StatusOK, items)
	}
}

// SessionSteps returns stored step timings for a session.
func (a *API) SessionSteps() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := strings.TrimSpace(r.PathValue("id"))
		if sessionID == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "sessionId is required"})
			return
		}
		steps, err := a.Store.SessionStepTimings(r.Context(), sessionID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, steps)
	}
}

// CompleteSession records a completed session and its step timings.
func (a *API) CompleteSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SessionID   string             `json:"sessionId"`
			WorkoutID   string             `json:"workoutId"`
			WorkoutName string             `json:"workoutName"`
			UserID      string             `json:"userId"`
			StartedAt   time.Time          `json:"startedAt"`
			CompletedAt time.Time          `json:"completedAt"`
			Steps       []SessionStepState `json:"steps"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.SessionID = strings.TrimSpace(req.SessionID)
		req.WorkoutID = strings.TrimSpace(req.WorkoutID)
		req.WorkoutName = strings.TrimSpace(req.WorkoutName)
		req.UserID = strings.TrimSpace(req.UserID)

		resolvedUserID, err := a.resolveUserID(r, req.UserID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
			return
		}
		req.UserID = resolvedUserID
		if req.SessionID == "" || req.WorkoutID == "" || req.UserID == "" {
			writeJSON(w, http.StatusBadRequest, apiError{Error: "sessionId, workoutId, and userId are required"})
			return
		}
		now := time.Now()
		if req.StartedAt.IsZero() {
			req.StartedAt = now
		}
		if req.CompletedAt.IsZero() {
			req.CompletedAt = now
		}
		if req.CompletedAt.Before(req.StartedAt) {
			req.CompletedAt = req.StartedAt.Add(time.Second)
		}
		var stepLogs []db.SessionStepLog
		for idx, st := range req.Steps {
			if st.ID == "" && st.Name == "" {
				continue
			}
			stepID := fmt.Sprintf("%s-%d", req.SessionID, idx)
			stepLogs = append(stepLogs, db.SessionStepLog{
				ID:               stepID,
				SessionID:        req.SessionID,
				StepOrder:        idx,
				Type:             strings.TrimSpace(st.Type),
				Name:             st.Name,
				EstimatedSeconds: st.EstimatedSeconds,
				ElapsedMillis:    st.ElapsedMillis,
			})
		}

		log := db.SessionLog{
			ID:          req.SessionID,
			WorkoutID:   req.WorkoutID,
			WorkoutName: req.WorkoutName,
			UserID:      req.UserID,
			StartedAt:   req.StartedAt,
			CompletedAt: req.CompletedAt,
		}
		if err := a.Store.RecordSession(r.Context(), log, stepLogs); err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, log)
	}
}

// buildSessionHistoryItems maps session logs to API response items.
func buildSessionHistoryItems(history []db.SessionLog, stepMap map[string][]db.SessionStepLog) []sessionHistoryItem {
	items := make([]sessionHistoryItem, 0, len(history))
	for _, h := range history {
		started := h.StartedAt
		completed := h.CompletedAt
		items = append(items, sessionHistoryItem{
			ID:          h.ID,
			SessionID:   h.ID,
			WorkoutID:   h.WorkoutID,
			WorkoutName: h.WorkoutName,
			UserID:      h.UserID,
			StartedAt:   &started,
			CompletedAt: &completed,
			Steps:       stepMap[h.ID],
		})
	}
	return items
}

// sessionStateFromWorkout builds an in-memory session state from a workout.
func sessionStateFromWorkout(workout *db.Workout) SessionState {
	state := SessionState{
		SessionID:    utils.NewID(),
		WorkoutID:    workout.ID,
		UserID:       workout.UserID,
		WorkoutName:  workout.Name,
		CurrentIndex: 0,
	}

	for idx, st := range workout.Steps {
		exercises := make([]Exercise, 0, len(st.Exercises))
		for _, ex := range st.Exercises {
			exercises = append(exercises, Exercise{
				Name:   ex.Name,
				Amount: ex.Amount,
				Weight: ex.Weight,
			})
		}
		stepState := SessionStepState{
			ID:               st.ID,
			Name:             st.Name,
			Type:             st.Type,
			EstimatedSeconds: st.EstimatedSeconds,
			SoundURL:         soundURLByKey(st.SoundKey),
			Exercises:        exercises,
			Current:          idx == 0,
		}
		autoAdvance := st.Type == "pause" && (strings.EqualFold(st.Weight, "__auto__") || st.PauseOptions.AutoAdvance)
		if autoAdvance {
			stepState.PauseOptions = PauseOptions{AutoAdvance: true}
		}
		state.Steps = append(state.Steps, stepState)
	}
	return state
}
