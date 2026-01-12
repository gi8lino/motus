package sessions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gi8lino/motus/internal/service"
	"github.com/gi8lino/motus/internal/utils"
)

// BuildSessionLog validates and maps a completion payload to a log and step entries.
func BuildSessionLog(req CompleteRequest) (SessionLog, []SessionStepLog, error) {
	req.SessionID = strings.TrimSpace(req.SessionID)
	req.WorkoutID = strings.TrimSpace(req.WorkoutID)
	req.WorkoutName = strings.TrimSpace(req.WorkoutName)
	req.UserID = strings.TrimSpace(req.UserID)

	if req.SessionID == "" || req.WorkoutID == "" || req.UserID == "" {
		return SessionLog{}, nil, service.NewError(service.ErrorValidation, "sessionId, workoutId, and userId are required")
	}
	// Default timestamps to now when missing.
	now := time.Now()
	req.StartedAt = utils.DefaultIfZero(req.StartedAt, now)
	req.CompletedAt = utils.DefaultIfZero(req.CompletedAt, now)

	// Ensure completion is never before the start.
	if req.CompletedAt.Before(req.StartedAt) {
		req.CompletedAt = req.StartedAt.Add(time.Second)
	}

	var stepLogs []SessionStepLog
	for idx, st := range req.Steps {
		// Skip empty entries emitted by the client.
		if st.ID == "" && st.Name == "" {
			continue
		}
		// Create a stable step log id per order position.
		stepID := fmt.Sprintf("%s-%d", req.SessionID, idx)
		stepLogs = append(stepLogs, SessionStepLog{
			ID:               stepID,
			SessionID:        req.SessionID,
			StepOrder:        idx,
			Type:             strings.TrimSpace(st.Type),
			Name:             strings.TrimSpace(st.Name),
			EstimatedSeconds: st.EstimatedSeconds,
			ElapsedMillis:    st.ElapsedMillis,
		})
	}

	log := SessionLog{
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
func RecordSession(ctx context.Context, store Store, req CompleteRequest) (SessionLog, error) {
	log, steps, err := BuildSessionLog(req)
	if err != nil {
		return SessionLog{}, err
	}

	if err := store.RecordSession(ctx, log, steps); err != nil {
		return SessionLog{}, service.NewError(service.ErrorInternal, err.Error())
	}

	return log, nil
}

// RecordSession persists a session log and its step timings.
func (s *Service) RecordSession(ctx context.Context, req CompleteRequest) (SessionLog, error) {
	return RecordSession(ctx, s.store, req)
}
