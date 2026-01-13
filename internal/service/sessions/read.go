package sessions

import (
	"context"
	"strings"

	domainSessions "github.com/gi8lino/motus/internal/domain/sessions"
	"github.com/gi8lino/motus/internal/service"
)

// SessionStateFromWorkout creates a session state by delegating to the session domain logic.
func SessionStateFromWorkout(workout *Workout, soundURLByKey func(string) string) SessionState {
	return domainSessions.NewStateFromWorkout(workout, soundURLByKey)
}

// CreateState builds a session state from a workout id.
func (s *Service) CreateState(ctx context.Context, workoutID string) (SessionState, error) {
	return CreateState(ctx, s.store, workoutID, s.soundURLByKey)
}

// CreateState builds a session state from a workout id.
func CreateState(ctx context.Context, store Store, workoutID string, soundURLByKey func(string) string) (SessionState, error) {
	workoutID = strings.TrimSpace(workoutID)
	if workoutID == "" {
		return SessionState{}, service.NewError(service.ErrorValidation, "workoutId is required")
	}
	workout, err := store.WorkoutWithSteps(ctx, workoutID)
	if err != nil {
		return SessionState{}, service.NewError(service.ErrorNotFound, err.Error())
	}
	return domainSessions.NewStateFromWorkout(workout, soundURLByKey), nil
}

// FetchStepTimings returns stored step timings for a session.
func (s *Service) FetchStepTimings(ctx context.Context, sessionID string) ([]SessionStepLog, error) {
	return FetchStepTimings(ctx, s.store, sessionID)
}

// FetchStepTimings returns stored step timings for a session.
func FetchStepTimings(ctx context.Context, store Store, sessionID string) ([]SessionStepLog, error) {
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

// BuildSessionHistory loads step timings and maps session logs to response items.
func (s *Service) BuildSessionHistory(ctx context.Context, userID string, limit int) ([]SessionHistoryItem, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, service.NewError(service.ErrorValidation, "userId is required")
	}
	history, err := s.store.SessionHistory(ctx, userID, limit)
	if err != nil {
		return nil, service.NewError(service.ErrorInternal, err.Error())
	}
	return BuildSessionHistory(ctx, s.store, history)
}

// BuildSessionHistory loads step timings and maps session logs to response items.
func BuildSessionHistory(ctx context.Context, store Store, history []SessionLog) ([]SessionHistoryItem, error) {
	// Collect step timing rows per session to enrich the history payload.
	stepMap := make(map[string][]SessionStepLog, len(history))
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
func BuildSessionHistoryItems(history []SessionLog, stepMap map[string][]SessionStepLog) []SessionHistoryItem {
	items := make([]SessionHistoryItem, 0, len(history))
	for _, h := range history {
		// Copy timestamps to avoid referencing loop variables.
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
