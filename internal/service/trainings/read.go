package trainings

import (
	"context"
	"strings"

	errpkg "github.com/gi8lino/motus/internal/service/errors"
)

// TrainingStateFromWorkout creates a training state by delegating to the training domain logic.
func TrainingStateFromWorkout(workout *Workout, soundURLByKey func(string) string) TrainingState {
	return NewStateFromWorkout(workout, soundURLByKey)
}

// CreateState builds a training state from a workout id.
func (s *Service) CreateState(ctx context.Context, workoutID string) (TrainingState, error) {
	return CreateState(ctx, s.store, workoutID, s.soundURLByKey)
}

// CreateState builds a training state from a workout id.
func CreateState(ctx context.Context, store Store, workoutID string, soundURLByKey func(string) string) (TrainingState, error) {
	workoutID = strings.TrimSpace(workoutID)
	if workoutID == "" {
		return TrainingState{}, errpkg.NewErrorWithScope(errpkg.ErrorValidation, "workoutId is required", errorScope)
	}
	workout, err := store.WorkoutWithSteps(ctx, workoutID)
	if err != nil {
		return TrainingState{}, errpkg.NewErrorWithScope(errpkg.ErrorNotFound, err.Error(), errorScope)
	}
	return NewStateFromWorkout(workout, soundURLByKey), nil
}

// FetchStepTimings returns stored step timings for a training.
func (s *Service) FetchStepTimings(ctx context.Context, trainingID string) ([]TrainingStepLog, error) {
	return FetchStepTimings(ctx, s.store, trainingID)
}

// FetchStepTimings returns stored step timings for a training.
func FetchStepTimings(ctx context.Context, store Store, trainingID string) ([]TrainingStepLog, error) {
	trainingID = strings.TrimSpace(trainingID)
	if trainingID == "" {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, "trainingId is required", errorScope)
	}

	steps, err := store.TrainingStepTimings(ctx, trainingID)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}

	return steps, nil
}

// BuildTrainingHistory loads step timings and maps training logs to response items.
func (s *Service) BuildTrainingHistory(ctx context.Context, userID string, limit int) ([]TrainingHistoryItem, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorValidation, "userId is required", errorScope)
	}
	history, err := s.store.TrainingHistory(ctx, userID, limit)
	if err != nil {
		return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
	}
	return BuildTrainingHistory(ctx, s.store, history)
}

// BuildTrainingHistory loads step timings and maps training logs to response items.
func BuildTrainingHistory(ctx context.Context, store Store, history []TrainingLog) ([]TrainingHistoryItem, error) {
	// Collect step timing rows per training to enrich the history payload.
	stepMap := make(map[string][]TrainingStepLog, len(history))
	for _, entry := range history {
		steps, err := store.TrainingStepTimings(ctx, entry.ID)
		if err != nil {
			return nil, errpkg.NewErrorWithScope(errpkg.ErrorInternal, err.Error(), errorScope)
		}
		stepMap[entry.ID] = steps
	}
	return BuildTrainingHistoryItems(history, stepMap), nil
}

// BuildTrainingHistoryItems maps training logs to API response items.
func BuildTrainingHistoryItems(history []TrainingLog, stepMap map[string][]TrainingStepLog) []TrainingHistoryItem {
	items := make([]TrainingHistoryItem, 0, len(history))
	for _, h := range history {
		// Copy timestamps to avoid referencing loop variables.
		started := h.StartedAt
		completed := h.CompletedAt
		items = append(items, TrainingHistoryItem{
			ID:          h.ID,
			TrainingID:  h.ID,
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
