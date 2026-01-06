package sessions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service"
	"github.com/gi8lino/motus/internal/utils"
)

// SessionHistoryItem is the API payload for a completed session.
type SessionHistoryItem struct {
	ID          string              `json:"id"`
	SessionID   string              `json:"sessionId"`
	WorkoutID   string              `json:"workoutId"`
	WorkoutName string              `json:"workoutName"`
	UserID      string              `json:"userId"`
	StartedAt   *time.Time          `json:"startedAt,omitempty"`
	CompletedAt *time.Time          `json:"completedAt,omitempty"`
	Steps       []db.SessionStepLog `json:"steps,omitempty"`
}

// SessionState mirrors the JSON contract with the SPA.
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

// SessionStepState describes a single step in a running session.
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

// Exercise describes a single exercise inside a step.
type Exercise struct {
	Name   string `json:"name"`
	Amount string `json:"amount"`
	Weight string `json:"weight"`
}

// PauseOptions describes pause behavior for session steps.
type PauseOptions struct {
	AutoAdvance bool `json:"autoAdvance"`
}

// store defines the persistence methods needed by the session helpers.
type store interface {
	SessionStepTimings(ctx context.Context, sessionID string) ([]db.SessionStepLog, error)
	WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error)
	RecordSession(ctx context.Context, log db.SessionLog, steps []db.SessionStepLog) error
}

// BuildSessionHistory loads step timings and maps session logs to response items.
func BuildSessionHistory(ctx context.Context, store store, history []db.SessionLog) ([]SessionHistoryItem, error) {
	stepMap := make(map[string][]db.SessionStepLog, len(history))
	for _, entry := range history {
		steps, err := store.SessionStepTimings(ctx, entry.ID)
		if err != nil {
			return nil, service.NewError(service.ErrorInternal, err.Error())
		}
		stepMap[entry.ID] = steps
	}
	return BuildSessionHistoryItems(history, stepMap), nil
}

// BuildSessionHistoryItems maps session logs to API response items.
func BuildSessionHistoryItems(history []db.SessionLog, stepMap map[string][]db.SessionStepLog) []SessionHistoryItem {
	items := make([]SessionHistoryItem, 0, len(history))
	for _, h := range history {
		started := h.StartedAt
		completed := h.CompletedAt
		items = append(items, SessionHistoryItem{
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

// SessionStateFromWorkout builds an in-memory session state from a workout.
func SessionStateFromWorkout(workout *db.Workout, soundURLByKey func(string) string) SessionState {
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

// CreateState builds a session state from a workout id.
func CreateState(ctx context.Context, store store, workoutID string, soundURLByKey func(string) string) (SessionState, error) {
	workoutID = strings.TrimSpace(workoutID)
	if workoutID == "" {
		return SessionState{}, service.NewError(service.ErrorValidation, "workoutId is required")
	}
	workout, err := store.WorkoutWithSteps(ctx, workoutID)
	if err != nil {
		return SessionState{}, service.NewError(service.ErrorNotFound, err.Error())
	}
	return SessionStateFromWorkout(workout, soundURLByKey), nil
}

// CompleteRequest captures the payload for logging a finished session.
type CompleteRequest struct {
	SessionID   string             `json:"sessionId"`
	WorkoutID   string             `json:"workoutId"`
	WorkoutName string             `json:"workoutName"`
	UserID      string             `json:"userId"`
	StartedAt   time.Time          `json:"startedAt"`
	CompletedAt time.Time          `json:"completedAt"`
	Steps       []SessionStepState `json:"steps"`
}

// BuildSessionLog validates and maps a completion payload to a log and step entries.
func BuildSessionLog(req CompleteRequest) (db.SessionLog, []db.SessionStepLog, error) {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.WorkoutID = strings.TrimSpace(req.WorkoutID)
	req.WorkoutName = strings.TrimSpace(req.WorkoutName)
	req.UserID = strings.TrimSpace(req.UserID)

	if req.SessionID == "" || req.WorkoutID == "" || req.UserID == "" {
		return db.SessionLog{}, nil, service.NewError(service.ErrorValidation, "sessionId, workoutId, and userId are required")
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
			Name:             strings.TrimSpace(st.Name),
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
	return log, stepLogs, nil
}

// RecordSession persists a session log and its step timings.
func RecordSession(ctx context.Context, store store, req CompleteRequest) (db.SessionLog, error) {
	log, steps, err := BuildSessionLog(req)
	if err != nil {
		return db.SessionLog{}, err
	}
	if err := store.RecordSession(ctx, log, steps); err != nil {
		return db.SessionLog{}, service.NewError(service.ErrorInternal, err.Error())
	}
	return log, nil
}

// FetchStepTimings returns stored step timings for a session.
func FetchStepTimings(ctx context.Context, store store, sessionID string) ([]db.SessionStepLog, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, service.NewError(service.ErrorValidation, "sessionId is required")
	}
	steps, err := store.SessionStepTimings(ctx, sessionID)
	if err != nil {
		return nil, service.NewError(service.ErrorInternal, err.Error())
	}
	return steps, nil
}
