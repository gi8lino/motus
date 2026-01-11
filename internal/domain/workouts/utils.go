package workouts

import (
	"strings"
	"time"
)

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
