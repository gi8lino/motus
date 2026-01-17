package workouts

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gi8lino/motus/internal/db"
	"github.com/gi8lino/motus/internal/utils"
)

var repRangePattern = regexp.MustCompile(`^\d+(-\d+)?$`)

// isNotFoundError reports whether the error matches a not found case.
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "not found")
}

// normalizeRepeatRest adjusts repeat rest metadata when there is at least one repeat.
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

// parseDurationField parses a duration string to seconds or returns the fallback.
func parseDurationField(value string, fallback int) (int, error) {
	if strings.TrimSpace(value) == "" {
		return max(fallback, 0), nil
	}
	dur, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, err
	}
	if dur < 0 {
		dur = 0
	}
	return int(dur / time.Second), nil
}

// isEmptyRepExercise returns true when a rep exercise has no meaningful payload.
func isEmptyRepExercise(ex ExerciseInput) bool {
	return strings.TrimSpace(ex.Name) == "" &&
		strings.TrimSpace(ex.Reps) == "" &&
		strings.TrimSpace(ex.Weight) == ""
}

// NormalizeSteps validates and converts step inputs into database steps.
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
			PauseOptions:          PauseOptions{AutoAdvance: autoAdvance},
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

// normalizeSubsets ensures every step contains at least one subset.
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

// normalizeSubset builds a WorkoutSubset, enforcing duration and sound rules.
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

// normalizeSubsetExercises converts exercise inputs while enforcing type-specific rules.
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
