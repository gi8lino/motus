package workouts

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service"
	"github.com/gi8lino/motus/internal/service/sounds"
	"github.com/gi8lino/motus/internal/utils"
)

// WorkoutRequest captures the payload for creating or updating workouts.
type WorkoutRequest struct {
	UserID string      `json:"userId"`
	Name   string      `json:"name"`
	Steps  []StepInput `json:"steps"`
}

// StepInput describes a workout step in an incoming request.
type StepInput struct {
	Type                  string          `json:"type"`
	Name                  string          `json:"name"`
	Duration              string          `json:"duration"`
	EstimatedSeconds      int             `json:"estimatedSeconds"`
	SoundKey              string          `json:"soundKey"`
	Exercises             []ExerciseInput `json:"exercises"`
	PauseOptions          db.PauseOptions `json:"pauseOptions"`
	RepeatCount           int             `json:"repeatCount"`
	RepeatRestSeconds     int             `json:"repeatRestSeconds"`
	RepeatRestAfterLast   bool            `json:"repeatRestAfterLast"`
	RepeatRestSoundKey    string          `json:"repeatRestSoundKey"`
	RepeatRestAutoAdvance bool            `json:"repeatRestAutoAdvance"`
}

// ExerciseInput describes a nested exercise entry in a step.
type ExerciseInput struct {
	ExerciseID string `json:"exerciseId"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Reps       string `json:"reps"`
	Weight     string `json:"weight"`
	Duration   string `json:"duration"`
}

var repRangePattern = regexp.MustCompile(`^\d+(-\d+)?$`)

// Store defines the persistence methods needed by the workouts service.
type Store interface {
	// CreateWorkout inserts a workout definition.
	CreateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error)
	// UpdateWorkout updates a workout definition.
	UpdateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error)
	// WorkoutsByUser returns workouts for a user.
	WorkoutsByUser(ctx context.Context, userID string) ([]db.Workout, error)
	// WorkoutWithSteps loads a workout and its steps.
	WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error)
	// DeleteWorkout removes a workout by id.
	DeleteWorkout(ctx context.Context, id string) error
}

// Service coordinates workout operations.
type Service struct {
	Store Store
}

// New creates a new workouts service.
func New(store Store) *Service {
	return &Service{Store: store}
}

// NormalizeSteps validates and normalizes workout steps from the request.
func NormalizeSteps(inputs []StepInput, validSoundKey func(string) bool) ([]db.WorkoutStep, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("at least one step is required")
	}

	steps := make([]db.WorkoutStep, 0, len(inputs))
	for idx := range inputs {
		in := inputs[idx]
		stepType := strings.TrimSpace(in.Type)
		name := strings.TrimSpace(in.Name)
		if stepType == "" || name == "" {
			return nil, fmt.Errorf("step %d requires name and type", idx+1)
		}
		if stepType != "set" && stepType != "pause" {
			return nil, fmt.Errorf("step %d has invalid type", idx+1)
		}

		// Parse duration if provided, otherwise fall back to estimated seconds.
		durationStr := strings.TrimSpace(in.Duration)
		seconds := in.EstimatedSeconds
		if durationStr != "" {
			dur, err := time.ParseDuration(durationStr)
			if err != nil {
				return nil, fmt.Errorf("invalid duration for %s: %w", name, err)
			}

			// Guard: avoid negative durations.
			dur = max(dur, 0)
			seconds = int(dur / time.Second)
		}

		// Guard: seconds cannot be negative.
		seconds = max(seconds, 0)
		soundKey := strings.TrimSpace(in.SoundKey)
		// Validate selected sound keys when provided.
		if validSoundKey != nil && !validSoundKey(soundKey) {
			return nil, fmt.Errorf("invalid sound selection for step %s", name)
		}

		repeatRestSoundKey := strings.TrimSpace(in.RepeatRestSoundKey)
		// Validate repeat rest sound if one was selected.
		if repeatRestSoundKey != "" && validSoundKey != nil && !validSoundKey(repeatRestSoundKey) {
			return nil, fmt.Errorf("invalid rest sound selection for step %s", name)
		}

		repeatCount := max(in.RepeatCount, 1) // Step always carries the actual repeat count. Must be at least 1 to enable rest
		repeatRestSeconds, repeatRestAutoAdvance, repeatRestAfterLast, repeatRestSoundKey := normalizeRepeatRest(
			repeatCount,
			max(in.RepeatRestSeconds, 0), // clamp to valid range (min 0)
			in.RepeatRestAutoAdvance,
			in.RepeatRestAfterLast,
			repeatRestSoundKey,
		)

		// Map pause auto-advance to the stored pause options.
		autoAdvance := stepType == "pause" && in.PauseOptions.AutoAdvance
		// Normalize exercises so blank rows are dropped.
		var exercises []db.StepExercise
		if len(in.Exercises) > 0 {
			for _, ex := range in.Exercises {
				exName := strings.TrimSpace(ex.Name)
				exType := utils.DefaultIfZero(utils.NormalizeToken(ex.Type), "rep")

				if exType != "rep" && exType != "timed" {
					return nil, fmt.Errorf("invalid exercise type for %s", name)
				}
				if exType == "timed" {
					durationText := strings.TrimSpace(ex.Duration)
					if durationText == "" {
						continue
					}
					if _, err := time.ParseDuration(durationText); err != nil {
						return nil, fmt.Errorf("invalid duration for %s", name)
					}
				}
				if exType == "rep" {
					repsText := strings.TrimSpace(ex.Reps)
					if repsText != "" && !repRangePattern.MatchString(repsText) {
						return nil, fmt.Errorf("invalid reps for %s", name)
					}
				}
				if exType == "rep" && isEmptyRepExercise(ex) {
					continue
				}

				exerciseID := strings.TrimSpace(ex.ExerciseID)
				reps := strings.TrimSpace(ex.Reps)
				weight := strings.TrimSpace(ex.Weight)
				duration := strings.TrimSpace(ex.Duration)
				if exType != "rep" {
					reps = ""
					weight = ""
				}
				if exType == "rep" {
					duration = ""
				}
				exercises = append(exercises, db.StepExercise{
					ExerciseID: exerciseID,
					Name:       exName,
					Type:       exType,
					Reps:       reps,
					Weight:     weight,
					Duration:   duration,
				})
			}
		}

		// Build the normalized step payload for storage.
		step := db.WorkoutStep{
			Type:                  stepType,
			Name:                  name,
			EstimatedSeconds:      seconds,
			SoundKey:              soundKey,
			Exercises:             exercises,
			PauseOptions:          db.PauseOptions{AutoAdvance: autoAdvance},
			RepeatCount:           repeatCount,
			RepeatRestSeconds:     repeatRestSeconds,
			RepeatRestAfterLast:   repeatRestAfterLast,
			RepeatRestSoundKey:    repeatRestSoundKey,
			RepeatRestAutoAdvance: repeatRestAutoAdvance,
		}
		if stepType == "pause" {
			// Pause steps shouldn't persist exercise rows; keep only the flag and duration.
			step.Exercises = nil
		}
		steps = append(steps, step)
	}
	return steps, nil
}

// Create stores a new workout for the user.
func (s *Service) Create(ctx context.Context, req WorkoutRequest) (*db.Workout, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.Name = strings.TrimSpace(req.Name)
	if req.UserID == "" || req.Name == "" || len(req.Steps) == 0 {
		return nil, service.NewError(service.ErrorValidation, "name and at least one step are required")
	}

	steps, err := NormalizeSteps(req.Steps, sounds.ValidKey)
	if err != nil {
		return nil, service.NewError(service.ErrorValidation, err.Error())
	}

	workout := &db.Workout{UserID: req.UserID, Name: req.Name, Steps: steps}
	created, err := s.Store.CreateWorkout(ctx, workout)
	if err != nil {
		return nil, service.NewError(service.ErrorInternal, err.Error())
	}

	return created, nil
}

// Update replaces a workout and its steps.
func (s *Service) Update(ctx context.Context, id string, req WorkoutRequest) (*db.Workout, error) {
	id = strings.TrimSpace(id)
	req.UserID = strings.TrimSpace(req.UserID)
	req.Name = strings.TrimSpace(req.Name)
	if id == "" {
		return nil, service.NewError(service.ErrorValidation, "workout id is required")
	}

	if req.Name == "" || len(req.Steps) == 0 {
		return nil, service.NewError(service.ErrorValidation, "name and steps are required")
	}

	steps, err := NormalizeSteps(req.Steps, sounds.ValidKey)
	if err != nil {
		return nil, service.NewError(service.ErrorValidation, err.Error())
	}

	workout := &db.Workout{ID: id, UserID: req.UserID, Name: req.Name, Steps: steps}
	updated, err := s.Store.UpdateWorkout(ctx, workout)
	if err != nil {
		return nil, service.NewError(service.ErrorInternal, err.Error())
	}

	return updated, nil
}

// Import creates a new workout from exported JSON.
func (s *Service) Import(ctx context.Context, userID string, workout db.Workout) (*db.Workout, error) {
	userID = strings.TrimSpace(userID)
	workout.Name = strings.TrimSpace(workout.Name)
	if userID == "" {
		return nil, service.NewError(service.ErrorValidation, "userId is required")
	}

	if workout.Name == "" || len(workout.Steps) == 0 {
		return nil, service.NewError(service.ErrorValidation, "workout name and steps are required")
	}

	// Reset ids and ordering so new rows are created.
	for idx := range workout.Steps {
		step := &workout.Steps[idx]
		step.ID = ""
		step.WorkoutID = ""
		step.Order = idx
		for exIdx := range step.Exercises {
			ex := &step.Exercises[exIdx]
			ex.ID = ""
			ex.StepID = ""
			ex.Order = exIdx
			ex.ExerciseID = ""
		}
	}
	created, err := s.Store.CreateWorkout(ctx, &db.Workout{
		UserID: userID,
		Name:   workout.Name,
		Steps:  workout.Steps,
	})
	if err != nil {
		return nil, service.NewError(service.ErrorInternal, err.Error())
	}

	return created, nil
}

// Get returns a workout by id.
func (s *Service) Get(ctx context.Context, id string) (*db.Workout, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, service.NewError(service.ErrorValidation, "workout id is required")
	}

	workout, err := s.Store.WorkoutWithSteps(ctx, id)
	if err != nil {
		return nil, service.NewError(service.ErrorNotFound, err.Error())
	}

	return workout, nil
}

// Export returns a workout for sharing.
func (s *Service) Export(ctx context.Context, id string) (*db.Workout, error) {
	return s.Get(ctx, id)
}

// Delete removes a workout by id.
func (s *Service) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return service.NewError(service.ErrorValidation, "workout id is required")
	}

	if err := s.Store.DeleteWorkout(ctx, id); err != nil {
		if err == pgx.ErrNoRows {
			return service.NewError(service.ErrorNotFound, "workout not found")
		}
		return service.NewError(service.ErrorInternal, err.Error())
	}

	return nil
}

// normalizeRepeatRest clears rest settings when repeats are disabled or rest seconds are zero.
func normalizeRepeatRest(
	repeatCount int,
	repeatRestSeconds int,
	repeatRestAutoAdvance bool,
	repeatRestAfterLast bool,
	repeatRestSoundKey string,
) (seconds int, autoAdvance bool, afterLast bool, soundKey string) {
	if repeatCount <= 1 || repeatRestSeconds == 0 {
		return 0, false, false, ""
	}
	return repeatRestSeconds, repeatRestAutoAdvance, repeatRestAfterLast, repeatRestSoundKey
}

// isEmptyRepExercise returns true when a rep exercise has no meaningful content.
func isEmptyRepExercise(ex ExerciseInput) bool {
	return strings.TrimSpace(ex.Name) == "" &&
		strings.TrimSpace(ex.Reps) == "" &&
		strings.TrimSpace(ex.Weight) == ""
}
