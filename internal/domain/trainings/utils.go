package trainings

import (
	"strings"
	"time"

	"github.com/gi8lino/motus/internal/utils"
)

// mapExercises converts DB subset exercises into training exercises.
func mapExercises(exercises []SubsetExercise) []Exercise {
	if len(exercises) == 0 {
		return nil
	}
	result := make([]Exercise, len(exercises))
	for i, ex := range exercises {
		result[i] = mapExercise(ex)
	}
	return result
}

// mapExercise builds a training exercise from a subset exercise record.
func mapExercise(ex SubsetExercise) Exercise {
	return Exercise{
		Name:     ex.Name,
		Type:     utils.NormalizeExerciseType(ex.Type),
		Reps:     ex.Reps,
		Weight:   ex.Weight,
		Duration: ex.Duration,
		SoundKey: ex.SoundKey,
	}
}

// deriveExerciseDuration determines timing for a subset exercise.
func deriveExerciseDuration(ex SubsetExercise, subset WorkoutSubset) (seconds int, autoAdvance bool) {
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

// parseDurationSeconds parses a duration string into seconds.
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
