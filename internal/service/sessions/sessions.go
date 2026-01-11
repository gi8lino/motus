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
	ID          string              `json:"id"`                    // ID is the history item identifier.
	SessionID   string              `json:"sessionId"`             // SessionID links to the completed session.
	WorkoutID   string              `json:"workoutId"`             // WorkoutID references the workout definition.
	WorkoutName string              `json:"workoutName"`           // WorkoutName is the display name at completion time.
	UserID      string              `json:"userId"`                // UserID owns the session.
	StartedAt   *time.Time          `json:"startedAt,omitempty"`   // StartedAt is when the session began.
	CompletedAt *time.Time          `json:"completedAt,omitempty"` // CompletedAt is when the session finished.
	Steps       []db.SessionStepLog `json:"steps,omitempty"`       // Steps contains logged timings when available.
}

// SessionState mirrors the JSON contract with the SPA.
type SessionState struct {
	SessionID    string             `json:"sessionId"`    // SessionID is the runtime identifier.
	WorkoutID    string             `json:"workoutId"`    // WorkoutID identifies the workout template.
	UserID       string             `json:"userId"`       // UserID owns the session.
	WorkoutName  string             `json:"workoutName"`  // WorkoutName is used for display.
	CurrentIndex int                `json:"currentIndex"` // CurrentIndex points to the active step.
	Running      bool               `json:"running"`      // Running reports whether the timer is active.
	Done         bool               `json:"done"`         // Done is true when the session is complete.
	StartedAt    time.Time          `json:"startedAt"`    // StartedAt records when the session started.
	CompletedAt  time.Time          `json:"completedAt"`  // CompletedAt records when the session finished.
	Steps        []SessionStepState `json:"steps"`        // Steps contains the ordered session steps.
}

// SessionStepState describes a single step in a running session.
type SessionStepState struct {
	ID                     string       `json:"id"`                               // ID is the unique session step id.
	Name                   string       `json:"name"`                             // Name is the step or exercise label.
	Type                   string       `json:"type"`                             // Type is the step kind (set or pause).
	EstimatedSeconds       int          `json:"estimatedSeconds"`                 // EstimatedSeconds is the per-step target time.
	SoundURL               string       `json:"soundUrl"`                         // SoundURL is the resolved audio path.
	SoundKey               string       `json:"soundKey,omitempty"`               // SoundKey references a bundled sound.
	SubsetEstimatedSeconds int          `json:"subsetEstimatedSeconds,omitempty"` // SubsetEstimatedSeconds targets the whole subset.
	Running                bool         `json:"running"`                          // Running reports whether this step is active.
	Completed              bool         `json:"completed"`                        // Completed reports whether this step finished.
	Current                bool         `json:"current"`                          // Current marks the active step.
	ElapsedMillis          int64        `json:"elapsedMillis"`                    // ElapsedMillis stores elapsed time for this step.
	Exercises              []Exercise   `json:"exercises"`                        // Exercises lists per-step exercises.
	PauseOptions           PauseOptions `json:"pauseOptions"`                     // PauseOptions configures pause behavior.
	AutoAdvance            bool         `json:"autoAdvance"`                      // AutoAdvance is true for timed auto-advance steps.
	LoopIndex              int          `json:"loopIndex,omitempty"`              // LoopIndex is the repeat cycle index.
	LoopTotal              int          `json:"loopTotal,omitempty"`              // LoopTotal is the repeat cycle count.

	SubsetID           string `json:"subsetId,omitempty"`           // SubsetID links the step to a subset definition.
	Superset           bool   `json:"superset,omitempty"`           // Superset marks subset blocks that skip to next subset.
	SubsetLabel        string `json:"subsetLabel,omitempty"`        // SubsetLabel is the subset display label.
	HasMultipleSubsets bool   `json:"hasMultipleSubsets,omitempty"` // HasMultipleSubsets flags grouped subsets.
	SetName            string `json:"setName,omitempty"`            // SetName is the parent set name.
}

// Exercise describes a single exercise inside a step.
type Exercise struct {
	Name     string `json:"name"`               // Name is the exercise name.
	Type     string `json:"type"`               // Type is the exercise kind.
	Reps     string `json:"reps"`               // Reps is the repetition count for rep exercises.
	Weight   string `json:"weight"`             // Weight is optional load text.
	Duration string `json:"duration"`           // Duration holds stopwatch/countdown text.
	SoundKey string `json:"soundKey,omitempty"` // SoundKey overrides the subset sound.
}

// PauseOptions describes pause behavior for session steps.
type PauseOptions struct {
	AutoAdvance bool `json:"autoAdvance"` // AutoAdvance skips to the next step on completion.
}

// Store defines the persistence methods needed by the session helpers.
type Store interface {
	// SessionStepTimings loads logged steps for a session.
	SessionStepTimings(ctx context.Context, sessionID string) ([]db.SessionStepLog, error)
	// WorkoutWithSteps loads a workout and its steps.
	WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error)
	// RecordSession stores a completed session and its steps.
	RecordSession(ctx context.Context, log db.SessionLog, steps []db.SessionStepLog) error
	// SessionHistory returns completed session logs for a user.
	SessionHistory(ctx context.Context, userID string, limit int) ([]db.SessionLog, error)
}

// BuildSessionHistory loads step timings and maps session logs to response items.
func BuildSessionHistory(ctx context.Context, store Store, history []db.SessionLog) ([]SessionHistoryItem, error) {
	// Collect step timing rows per session to enrich the history payload.
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

// SessionStateFromWorkout builds an in-memory session state from a workout.
func SessionStateFromWorkout(workout *db.Workout, soundURLByKey func(string) string) SessionState {
	state := SessionState{
		SessionID:    utils.NewID(),
		WorkoutID:    workout.ID,
		UserID:       workout.UserID,
		WorkoutName:  workout.Name,
		CurrentIndex: 0,
	}

	for _, st := range workout.Steps {
		repeatCount := max(st.RepeatCount, 1)
		hasMultipleSubsets := len(st.Subsets) > 1
		for loopIdx := range repeatCount {
			idBase := st.ID
			if repeatCount > 1 {
				idBase = fmt.Sprintf("%s-r%d", st.ID, loopIdx+1)
			}

			if st.Type == utils.StepTypePause.String() {
				pauseState := SessionStepState{
					ID:               idBase,
					Name:             st.Name,
					Type:             utils.StepTypePause.String(),
					EstimatedSeconds: st.EstimatedSeconds,
					SoundURL:         soundURLByKey(st.SoundKey),
					Current:          len(state.Steps) == 0,
					SetName:          st.Name,
				}
				if st.PauseOptions.AutoAdvance {
					pauseState.PauseOptions = PauseOptions{AutoAdvance: true}
				}
				if repeatCount > 1 {
					pauseState.LoopIndex = loopIdx + 1
					pauseState.LoopTotal = repeatCount
				}
				state.Steps = append(state.Steps, pauseState)
				continue
			}

			for subsetIdx := range st.Subsets {
				sub := st.Subsets[subsetIdx]
				subsetID := sub.ID
				subsetBase := fmt.Sprintf("%s-sub-%d", idBase, subsetIdx+1)
				subsetLabel := strings.TrimSpace(sub.Name)
				if sub.Superset {
					stepState := SessionStepState{
						ID:                     fmt.Sprintf("%s", subsetBase),
						Name:                   sub.Name,
						Type:                   st.Type,
						EstimatedSeconds:       sub.EstimatedSeconds,
						SoundURL:               soundURLByKey(sub.SoundKey),
						SoundKey:               sub.SoundKey,
						Exercises:              mapExercises(sub.Exercises),
						Current:                len(state.Steps) == 0,
						Superset:               true,
						SubsetID:               subsetID,
						SubsetLabel:            subsetLabel,
						HasMultipleSubsets:     hasMultipleSubsets,
						SetName:                st.Name,
						SubsetEstimatedSeconds: sub.EstimatedSeconds,
					}
					if repeatCount > 1 {
						stepState.LoopIndex = loopIdx + 1
						stepState.LoopTotal = repeatCount
					}
					state.Steps = append(state.Steps, stepState)
					continue
				}

				for exIdx, ex := range sub.Exercises {
					stepID := fmt.Sprintf("%s-ex-%d", subsetBase, exIdx+1)
					estimatedSeconds, autoAdvance := deriveExerciseDuration(ex, sub)
					stepState := SessionStepState{
						ID:                     stepID,
						Name:                   ex.Name,
						Type:                   st.Type,
						EstimatedSeconds:       estimatedSeconds,
						SoundURL:               soundURLByKey(sub.SoundKey),
						SoundKey:               sub.SoundKey,
						Exercises:              []Exercise{mapExercise(ex)},
						Current:                len(state.Steps) == 0,
						SubsetID:               subsetID,
						SubsetLabel:            subsetLabel,
						HasMultipleSubsets:     hasMultipleSubsets,
						SetName:                st.Name,
						SubsetEstimatedSeconds: sub.EstimatedSeconds,
						AutoAdvance:            autoAdvance,
					}
					if repeatCount > 1 {
						stepState.LoopIndex = loopIdx + 1
						stepState.LoopTotal = repeatCount
					}
					state.Steps = append(state.Steps, stepState)
				}
			}

			if st.RepeatRestSeconds > 0 && (loopIdx < repeatCount-1 || st.RepeatRestAfterLast) {
				restState := SessionStepState{
					ID:               fmt.Sprintf("%s-rest-%d", st.ID, loopIdx+1),
					Name:             "Pause",
					Type:             utils.StepTypePause.String(),
					EstimatedSeconds: st.RepeatRestSeconds,
					SoundURL:         soundURLByKey(st.RepeatRestSoundKey),
					Current:          len(state.Steps) == 0,
					SetName:          "Pause",
				}
				if st.RepeatRestAutoAdvance {
					restState.PauseOptions = PauseOptions{AutoAdvance: true}
				}
				if repeatCount > 1 {
					restState.LoopIndex = loopIdx + 1
					restState.LoopTotal = repeatCount
				}
				state.Steps = append(state.Steps, restState)
			}
		}
	}
	return state
}

func mapExercises(exercises []db.SubsetExercise) []Exercise {
	if len(exercises) == 0 {
		return nil
	}
	result := make([]Exercise, len(exercises))
	for i, ex := range exercises {
		result[i] = mapExercise(ex)
	}
	return result
}

func mapExercise(ex db.SubsetExercise) Exercise {
	return Exercise{
		Name:     ex.Name,
		Type:     utils.NormalizeExerciseType(ex.Type),
		Reps:     ex.Reps,
		Weight:   ex.Weight,
		Duration: ex.Duration,
		SoundKey: ex.SoundKey,
	}
}

func deriveExerciseDuration(
	ex db.SubsetExercise,
	subset db.WorkoutSubset,
) (seconds int, autoAdvance bool) {
	exType := utils.NormalizeExerciseType(ex.Type)
	if exType == utils.ExerciseTypeCountdown || exType == utils.ExerciseTypeStopwatch {
		dur := parseDurationSeconds(ex.Duration)
		if dur <= 0 && subset.EstimatedSeconds > 0 {
			dur = subset.EstimatedSeconds
		}
		return dur, exType == utils.ExerciseTypeCountdown && dur > 0
	}
	if exType == utils.ExerciseTypeRep && len(subset.Exercises) == 1 && subset.EstimatedSeconds > 0 {
		return subset.EstimatedSeconds, false
	}
	return 0, false
}

func parseDurationSeconds(value string) int {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		if dur, err := time.ParseDuration(trimmed); err == nil {
			if dur < 0 {
				return 0
			}
			return int(dur / time.Second)
		}
	}
	return 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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
	return SessionStateFromWorkout(workout, soundURLByKey), nil
}

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

// FetchStepTimings returns stored step timings for a session.
func FetchStepTimings(ctx context.Context, store Store, sessionID string) ([]db.SessionStepLog, error) {
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
