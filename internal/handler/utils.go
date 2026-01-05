package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/gi8lino/motus/internal/db"
)

type workoutRequest struct {
	UserID string      `json:"userId"`
	Name   string      `json:"name"`
	Steps  []stepInput `json:"steps"`
}

type stepInput struct {
	Type             string          `json:"type"`
	Name             string          `json:"name"`
	Duration         string          `json:"duration"`
	EstimatedSeconds int             `json:"estimatedSeconds"`
	SoundKey         string          `json:"soundKey"`
	Exercise         string          `json:"exercise"`
	Amount           string          `json:"amount"`
	Weight           string          `json:"weight"`
	Exercises        []exerciseInput `json:"exercises"`
	PauseOptions     PauseOptions    `json:"pauseOptions"`
}

type exerciseInput struct {
	ExerciseID string `json:"exerciseId"`
	Name       string `json:"name"`
	Amount     string `json:"amount"`
	Weight     string `json:"weight"`
}

// normalizeSteps validates and normalizes workout steps from the request.
func normalizeSteps(inputs []stepInput) ([]db.WorkoutStep, error) {
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
			if dur < 0 {
				dur = 0
			}
			seconds = int(dur / time.Second)
		}
		if seconds < 0 {
			seconds = 0
		}
		soundKey := strings.TrimSpace(in.SoundKey)
		if !validSoundKey(soundKey) {
			return nil, fmt.Errorf("invalid sound selection for step %s", name)
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
			Type:             stepType,
			Name:             name,
			EstimatedSeconds: seconds,
			SoundKey:         soundKey,
			Exercise:         strings.TrimSpace(in.Exercise),
			Amount:           strings.TrimSpace(in.Amount),
			Weight:           weight,
			Exercises:        exercises,
			PauseOptions:     db.PauseOptions{AutoAdvance: autoAdvance},
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

type soundOption struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	File        string `json:"file"`
	LeadSeconds int    `json:"leadSeconds,omitempty"`
}

var builtinSoundOptions = []soundOption{
	{Key: "beep", Label: "Soft Beep", File: "/sounds/beep.wav"},
	{Key: "chime", Label: "Gentle Chime", File: "/sounds/chime.wav"},
	{Key: "click", Label: "Click", File: "/sounds/click.wav"},
	{Key: "soft1", Label: "Friendly Soft 1", File: "/sounds/soft1.wav"},
	{Key: "soft2", Label: "Friendly Soft 2", File: "/sounds/soft2.wav"},
	{Key: "soft3", Label: "Friendly Soft 3", File: "/sounds/soft3.wav"},
	{Key: "soft4", Label: "Friendly Soft 4", File: "/sounds/soft4.wav"},
	{Key: "countdown", Label: "Countdown Tut-Tut-Tuuu", File: "/sounds/countdown.wav"},
	{Key: "race", Label: "Race Start", File: "/sounds/race.wav"},
	{Key: "count321", Label: "3-2-1 Action", File: "/sounds/count321.wav", LeadSeconds: 3},
}

var soundLookup = func() map[string]soundOption {
	m := make(map[string]soundOption, len(builtinSoundOptions))
	for _, opt := range builtinSoundOptions {
		m[opt.Key] = opt
	}
	return m
}()

// soundURLByKey resolves a sound key to its file path.
func soundURLByKey(key string) string {
	if opt, ok := soundLookup[key]; ok {
		return opt.File
	}
	return ""
}

// validSoundKey reports whether a sound key is recognized.
func validSoundKey(key string) bool {
	if key == "" {
		return true
	}
	_, ok := soundLookup[key]
	return ok
}
