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
	UserID string      `json:"userId"` // UserID owns the workout.
	Name   string      `json:"name"`   // Name is the workout title.
	Steps  []StepInput `json:"steps"`  // Steps defines the workout flow.
}

// StepInput describes a workout step in an incoming request.
type StepInput struct {
	Type                  string          `json:"type"`                  // Type is set or pause.
	Name                  string          `json:"name"`                  // Name is the step label.
	Duration              string          `json:"duration"`              // Duration is the pause target text.
	EstimatedSeconds      int             `json:"estimatedSeconds"`      // EstimatedSeconds stores a parsed target time.
	SoundKey              string          `json:"soundKey"`              // SoundKey plays on step completion.
	Subsets               []SubsetInput   `json:"subsets"`               // Subsets hold exercises for set steps.
	PauseOptions          db.PauseOptions `json:"pauseOptions"`          // PauseOptions configure pause behavior.
	RepeatCount           int             `json:"repeatCount"`           // RepeatCount repeats the step group.
	RepeatRestSeconds     int             `json:"repeatRestSeconds"`     // RepeatRestSeconds is rest duration between repeats.
	RepeatRestAfterLast   bool            `json:"repeatRestAfterLast"`   // RepeatRestAfterLast includes a rest after the last repeat.
	RepeatRestSoundKey    string          `json:"repeatRestSoundKey"`    // RepeatRestSoundKey plays during repeat rest.
	RepeatRestAutoAdvance bool            `json:"repeatRestAutoAdvance"` // RepeatRestAutoAdvance skips rests automatically.
}

// ExerciseInput describes a nested exercise entry in a step.
type ExerciseInput struct {
	ExerciseID string `json:"exerciseId"` // ExerciseID links to the catalog.
	Name       string `json:"name"`       // Name is the exercise label.
	Type       string `json:"type"`       // Type is rep, stopwatch, or countdown.
	Reps       string `json:"reps"`       // Reps is the repetition text.
	Weight     string `json:"weight"`     // Weight is optional load text.
	Duration   string `json:"duration"`   // Duration is a stopwatch/countdown value.
	SoundKey   string `json:"soundKey"`   // SoundKey overrides the subset sound.
}

// SubsetInput describes a logical subset inside a set step.
type SubsetInput struct {
	Name      string          `json:"name"`      // Name is an optional subset label.
	Duration  string          `json:"duration"`  // Duration sets the subset target time.
	SoundKey  string          `json:"soundKey"`  // SoundKey plays at the subset target.
	Superset  bool            `json:"superset"`  // Superset skips to the next subset on Next.
	Exercises []ExerciseInput `json:"exercises"` // Exercises belong to the subset.
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
		rawType := strings.TrimSpace(in.Type)
		name := strings.TrimSpace(in.Name)
		if rawType == "" || name == "" {
			return nil, fmt.Errorf("step %d requires name and type", idx+1)
		}
		if rawType != utils.StepTypeSet.String() && rawType != utils.StepTypePause.String() {
			return nil, fmt.Errorf("step %d has invalid type", idx+1)
		}
		stepType := utils.NormalizeStepType(rawType)

		durationSeconds, err := parseDurationField(in.Duration, in.EstimatedSeconds)
		if err != nil {
			return nil, fmt.Errorf("invalid duration for %s: %w", name, err)
		}

		soundKey := strings.TrimSpace(in.SoundKey)
		if validSoundKey != nil && soundKey != "" && !validSoundKey(soundKey) {
			return nil, fmt.Errorf("invalid sound selection for step %s", name)
		}

		repeatRestSoundKey := strings.TrimSpace(in.RepeatRestSoundKey)
		if repeatRestSoundKey != "" && validSoundKey != nil && !validSoundKey(repeatRestSoundKey) {
			return nil, fmt.Errorf("invalid rest sound selection for step %s", name)
		}

		repeatCount := max(in.RepeatCount, 1)
		repeatRestSeconds, repeatRestAutoAdvance, repeatRestAfterLast, repeatRestSoundKey := normalizeRepeatRest(
			repeatCount,
			max(in.RepeatRestSeconds, 0),
			in.RepeatRestAutoAdvance,
			in.RepeatRestAfterLast,
			repeatRestSoundKey,
		)

		autoAdvance := stepType == utils.StepTypePause && in.PauseOptions.AutoAdvance

		step := db.WorkoutStep{
			Type:                  stepType.String(),
			Name:                  name,
			EstimatedSeconds:      0,
			SoundKey:              soundKey,
			PauseOptions:          db.PauseOptions{AutoAdvance: autoAdvance},
			RepeatCount:           repeatCount,
			RepeatRestSeconds:     repeatRestSeconds,
			RepeatRestAfterLast:   repeatRestAfterLast,
			RepeatRestSoundKey:    repeatRestSoundKey,
			RepeatRestAutoAdvance: repeatRestAutoAdvance,
		}

		if stepType == utils.StepTypePause {
			step.EstimatedSeconds = durationSeconds
		} else {
			subsets, err := normalizeSubsets(name, in.Subsets, validSoundKey)
			if err != nil {
				return nil, err
			}
			step.Subsets = subsets
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
		for subIdx := range step.Subsets {
			sub := &step.Subsets[subIdx]
			sub.ID = ""
			sub.StepID = ""
			sub.Order = subIdx
			for exIdx := range sub.Exercises {
				ex := &sub.Exercises[exIdx]
				ex.ID = ""
				ex.SubsetID = ""
				ex.Order = exIdx
				ex.ExerciseID = ""
			}
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
func parseDurationField(value string, fallback int) (int, error) {
	if strings.TrimSpace(value) == "" {
		return max(fallback, 0), nil
	}
	dur, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, err
	}
	dur = max(dur, 0)
	return int(dur / time.Second), nil
}

func normalizeSubsets(stepName string, inputs []SubsetInput, validSoundKey func(string) bool) ([]db.WorkoutSubset, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("%s requires at least one subset", stepName)
	}
	subsets := make([]db.WorkoutSubset, 0, len(inputs))
	for idx := range inputs {
		subset, err := normalizeSubset(stepName, idx, inputs[idx], validSoundKey)
		if err != nil {
			return nil, err
		}
		subsets = append(subsets, subset)
	}
	return subsets, nil
}

func normalizeSubset(stepName string, index int, input SubsetInput, validSoundKey func(string) bool) (db.WorkoutSubset, error) {
	name := strings.TrimSpace(input.Name)
	label := name
	if label == "" {
		label = fmt.Sprintf("subset %d of %s", index+1, stepName)
	}
	seconds, err := parseDurationField(input.Duration, 0)
	if err != nil {
		return db.WorkoutSubset{}, fmt.Errorf("invalid duration for %s: %w", label, err)
	}
	soundKey := strings.TrimSpace(input.SoundKey)
	if validSoundKey != nil && soundKey != "" && !validSoundKey(soundKey) {
		return db.WorkoutSubset{}, fmt.Errorf("invalid sound for %s", label)
	}
	exercises, err := normalizeSubsetExercises(label, input.Exercises, validSoundKey)
	if err != nil {
		return db.WorkoutSubset{}, err
	}
	return db.WorkoutSubset{
		Name:             name,
		EstimatedSeconds: seconds,
		SoundKey:         soundKey,
		Superset:         input.Superset,
		Exercises:        exercises,
	}, nil
}

func normalizeSubsetExercises(name string, inputs []ExerciseInput, validSoundKey func(string) bool) ([]db.SubsetExercise, error) {
	var exercises []db.SubsetExercise
	for _, ex := range inputs {
		exName := strings.TrimSpace(ex.Name)
		token := utils.NormalizeToken(ex.Type)
		if token == "" {
			token = utils.ExerciseTypeRep
		}
		switch token {
		case utils.ExerciseTypeRep, utils.ExerciseTypeStopwatch, utils.ExerciseTypeCountdown:
		default:
			return nil, fmt.Errorf("invalid exercise type for %s", name)
		}
		exType := utils.NormalizeExerciseType(token)
		if exType == utils.ExerciseTypeCountdown || exType == utils.ExerciseTypeStopwatch {
			durationText := strings.TrimSpace(ex.Duration)
			if exType == utils.ExerciseTypeCountdown && durationText == "" {
				return nil, fmt.Errorf("invalid duration for %s", name)
			}
			if durationText != "" {
				if _, err := time.ParseDuration(durationText); err != nil {
					return nil, fmt.Errorf("invalid duration for %s", name)
				}
			}
		}
		if exType == utils.ExerciseTypeRep {
			repsText := strings.TrimSpace(ex.Reps)
			if repsText != "" && !repRangePattern.MatchString(repsText) {
				return nil, fmt.Errorf("invalid reps for %s", name)
			}
		}
		if exType == utils.ExerciseTypeRep && isEmptyRepExercise(ex) {
			continue
		}

		exerciseID := strings.TrimSpace(ex.ExerciseID)
		reps := strings.TrimSpace(ex.Reps)
		weight := strings.TrimSpace(ex.Weight)
		duration := strings.TrimSpace(ex.Duration)
		soundKey := strings.TrimSpace(ex.SoundKey)
		if validSoundKey != nil && soundKey != "" && !validSoundKey(soundKey) {
			return nil, fmt.Errorf("invalid exercise sound for %s", name)
		}
		if exType != utils.ExerciseTypeRep {
			reps = ""
		}
		if exType == utils.ExerciseTypeRep {
			duration = ""
		}
		exercises = append(exercises, db.SubsetExercise{
			ExerciseID: exerciseID,
			Name:       exName,
			Type:       exType,
			Reps:       reps,
			Weight:     weight,
			Duration:   duration,
			SoundKey:   soundKey,
		})
	}
	if len(exercises) == 0 {
		return nil, fmt.Errorf("subset %s requires at least one exercise", name)
	}
	return exercises, nil
}

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
