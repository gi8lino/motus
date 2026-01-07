package workouts

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/service"
	"github.com/gi8lino/motus/internal/service/sounds"
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
	Exercise              string          `json:"exercise"`
	Amount                string          `json:"amount"`
	Weight                string          `json:"weight"`
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
	Amount     string `json:"amount"`
	Weight     string `json:"weight"`
}

// store defines the persistence methods needed by the workouts service.
type store interface {
	CreateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error)
	UpdateWorkout(ctx context.Context, workout *db.Workout) (*db.Workout, error)
	WorkoutsByUser(ctx context.Context, userID string) ([]db.Workout, error)
	WorkoutWithSteps(ctx context.Context, id string) (*db.Workout, error)
	DeleteWorkout(ctx context.Context, id string) error
}

// Service coordinates workout operations.
type Service struct {
	Store store
}

// New creates a new workouts service.
func New(store store) *Service {
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
		if stepType == "timed" {
			// Timed steps use per-exercise durations; keep step duration zero.
			in.Duration = ""
			in.EstimatedSeconds = 0
		}
		durationStr := strings.TrimSpace(in.Duration)
		seconds := in.EstimatedSeconds
		if durationStr != "" {
			dur, err := time.ParseDuration(durationStr)
			if err != nil {
				return nil, fmt.Errorf("invalid duration for %s: %w", name, err)
			}
			dur = max(dur, 0)
			seconds = int(dur / time.Second)
		}
		seconds = max(seconds, 0)
		soundKey := strings.TrimSpace(in.SoundKey)
		if validSoundKey != nil && !validSoundKey(soundKey) {
			return nil, fmt.Errorf("invalid sound selection for step %s", name)
		}

		repeatCount := max(in.RepeatCount, 1)
		repeatRestSeconds := max(in.RepeatRestSeconds, 0)

		repeatRestSoundKey := strings.TrimSpace(in.RepeatRestSoundKey)
		if repeatRestSoundKey != "" && validSoundKey != nil && !validSoundKey(repeatRestSoundKey) {
			return nil, fmt.Errorf("invalid rest sound selection for step %s", name)
		}
		repeatRestAutoAdvance := in.RepeatRestAutoAdvance
		repeatRestAfterLast := in.RepeatRestAfterLast
		if repeatCount <= 1 {
			repeatRestSeconds = 0
			repeatRestAutoAdvance = false
			repeatRestAfterLast = false
			repeatRestSoundKey = ""
		}
		if repeatRestSeconds == 0 {
			repeatRestAutoAdvance = false
			repeatRestAfterLast = false
			repeatRestSoundKey = ""
		}
		autoAdvance := stepType == "pause" && in.PauseOptions.AutoAdvance
		weight := strings.TrimSpace(in.Weight)
		if autoAdvance {
			weight = "__auto__"
		}
		if stepType == "pause" && !autoAdvance && strings.EqualFold(weight, "__auto__") {
			weight = ""
		}
		var exercises []db.StepExercise
		if len(in.Exercises) > 0 {
			for _, ex := range in.Exercises {
				exName := strings.TrimSpace(ex.Name)
				if exName == "" && strings.TrimSpace(ex.Amount) == "" && strings.TrimSpace(ex.Weight) == "" {
					continue
				}
				exercises = append(exercises, db.StepExercise{
					ExerciseID: strings.TrimSpace(ex.ExerciseID),
					Name:       exName,
					Amount:     strings.TrimSpace(ex.Amount),
					Weight:     strings.TrimSpace(ex.Weight),
				})
			}
		} else if strings.TrimSpace(in.Exercise) != "" || strings.TrimSpace(in.Amount) != "" || strings.TrimSpace(in.Weight) != "" {
			exercises = append(exercises, db.StepExercise{
				Name:   strings.TrimSpace(in.Exercise),
				Amount: strings.TrimSpace(in.Amount),
				Weight: strings.TrimSpace(in.Weight),
			})
		}

		step := db.WorkoutStep{
			Type:                  stepType,
			Name:                  name,
			EstimatedSeconds:      seconds,
			SoundKey:              soundKey,
			Exercise:              strings.TrimSpace(in.Exercise),
			Amount:                strings.TrimSpace(in.Amount),
			Weight:                weight,
			Exercises:             exercises,
			PauseOptions:          db.PauseOptions{AutoAdvance: autoAdvance},
			RepeatCount:           repeatCount,
			RepeatRestSeconds:     repeatRestSeconds,
			RepeatRestAfterLast:   repeatRestAfterLast,
			RepeatRestSoundKey:    repeatRestSoundKey,
			RepeatRestAutoAdvance: repeatRestAutoAdvance,
		}
		if len(exercises) > 0 && stepType != "pause" {
			step.Exercise = exercises[0].Name
			step.Amount = exercises[0].Amount
			step.Weight = exercises[0].Weight
		}
		if stepType == "pause" {
			// Pause steps shouldn't persist exercise rows; keep only the flag and duration.
			step.Exercises = nil
			step.Exercise = ""
			step.Amount = ""
			if !autoAdvance && strings.EqualFold(step.Weight, "__auto__") {
				step.Weight = ""
			}
		}
		steps = append(steps, step)
	}
	return steps, nil
}

// Create stores a new workout for the user.
func (s *Service) Create(ctx context.Context, req WorkoutRequest) (*db.Workout, error) {
	req.UserID = strings.TrimSpace(req.UserID)
	req.Name = strings.TrimSpace(req.Name)
	if req.UserID == "" || strings.TrimSpace(req.Name) == "" || len(req.Steps) == 0 {
		return nil, service.NewError(service.ErrorValidation, "name and at least one step are required")
	}
	steps, err := NormalizeSteps(req.Steps, sounds.ValidKey)
	if err != nil {
		return nil, service.NewError(service.ErrorValidation, err.Error())
	}
	workout := &db.Workout{UserID: req.UserID, Name: strings.TrimSpace(req.Name), Steps: steps}
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
	if strings.TrimSpace(req.Name) == "" || len(req.Steps) == 0 {
		return nil, service.NewError(service.ErrorValidation, "name and steps are required")
	}
	steps, err := NormalizeSteps(req.Steps, sounds.ValidKey)
	if err != nil {
		return nil, service.NewError(service.ErrorValidation, err.Error())
	}
	workout := &db.Workout{ID: id, UserID: req.UserID, Name: strings.TrimSpace(req.Name), Steps: steps}
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
