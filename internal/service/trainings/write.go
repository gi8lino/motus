package trainings

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gi8lino/motus/internal/service"
	"github.com/gi8lino/motus/internal/utils"
)

// BuildTrainingLog validates and maps a completion payload to a log and step entries.
func BuildTrainingLog(req CompleteRequest) (TrainingLog, []TrainingStepLog, error) {
	req.TrainingID = strings.TrimSpace(req.TrainingID)
	req.WorkoutID = strings.TrimSpace(req.WorkoutID)
	req.WorkoutName = strings.TrimSpace(req.WorkoutName)
	req.UserID = strings.TrimSpace(req.UserID)

	if req.TrainingID == "" || req.WorkoutID == "" || req.UserID == "" {
		return TrainingLog{}, nil, service.NewError(service.ErrorValidation, "trainingId, workoutId, and userId are required")
	}
	// Default timestamps to now when missing.
	now := time.Now()
	req.StartedAt = utils.DefaultIfZero(req.StartedAt, now)
	req.CompletedAt = utils.DefaultIfZero(req.CompletedAt, now)

	// Ensure completion is never before the start.
	if req.CompletedAt.Before(req.StartedAt) {
		req.CompletedAt = req.StartedAt.Add(time.Second)
	}

	var stepLogs []TrainingStepLog
	for idx, st := range req.Steps {
		// Skip empty entries emitted by the client.
		if st.ID == "" && st.Name == "" {
			continue
		}
		// Create a stable step log id per order position.
		stepID := fmt.Sprintf("%s-%d", req.TrainingID, idx)
		stepLogs = append(stepLogs, TrainingStepLog{
			ID:               stepID,
			TrainingID:       req.TrainingID,
			StepOrder:        idx,
			Type:             strings.TrimSpace(st.Type),
			Name:             strings.TrimSpace(st.Name),
			EstimatedSeconds: st.EstimatedSeconds,
			ElapsedMillis:    st.ElapsedMillis,
		})
	}

	log := TrainingLog{
		ID:          req.TrainingID,
		WorkoutID:   req.WorkoutID,
		WorkoutName: req.WorkoutName,
		UserID:      req.UserID,
		StartedAt:   req.StartedAt,
		CompletedAt: req.CompletedAt,
	}

	return log, stepLogs, nil
}

// RecordTraining persists a training log and its step timings.
func RecordTraining(ctx context.Context, store Store, req CompleteRequest) (TrainingLog, error) {
	log, steps, err := BuildTrainingLog(req)
	if err != nil {
		return TrainingLog{}, err
	}

	if err := store.RecordTraining(ctx, log, steps); err != nil {
		return TrainingLog{}, service.NewError(service.ErrorInternal, err.Error())
	}

	return log, nil
}

// RecordTraining persists a training log and its step timings.
func (s *Service) RecordTraining(ctx context.Context, req CompleteRequest) (TrainingLog, error) {
	return RecordTraining(ctx, s.store, req)
}
