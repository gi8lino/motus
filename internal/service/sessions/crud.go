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

// CompleteRequest captures the payload for logging a finished session.
type CompleteRequest struct {
	SessionID   string             `json:"sessionId"`   // SessionID identifies the session.
	WorkoutID   string             `json:"workoutId"`   // WorkoutID identifies the workout.
	WorkoutName string             `json:"workoutName"` // WorkoutName is the display name at completion time.
	UserID      string             `json:"userId"`      // UserID owns the session.
	StartedAt   time.Time          `json:"startedAt"`   // StartedAt records when the session began.
	CompletedAt time.Time          `json:"completedAt"` // CompletedAt records when the session finished.
	Steps       []SessionStepState `json:"steps"`       // Steps includes timing details.
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
	// Default timestamps to now when missing.
	now := time.Now()
	req.StartedAt = utils.DefaultIfZero(req.StartedAt, now)
	req.CompletedAt = utils.DefaultIfZero(req.CompletedAt, now)

	// Ensure completion is never before the start.
	if req.CompletedAt.Before(req.StartedAt) {
		req.CompletedAt = req.StartedAt.Add(time.Second)
	}

	var stepLogs []db.SessionStepLog
	for idx, st := range req.Steps {
		// Skip empty entries emitted by the client.
		if st.ID == "" && st.Name == "" {
			continue
		}
		// Create a stable step log id per order position.
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
func RecordSession(ctx context.Context, store Store, req CompleteRequest) (db.SessionLog, error) {
	log, steps, err := BuildSessionLog(req)
	if err != nil {
		return db.SessionLog{}, err
	}

	if err := store.RecordSession(ctx, log, steps); err != nil {
		return db.SessionLog{}, service.NewError(service.ErrorInternal, err.Error())
	}

	return log, nil
}
