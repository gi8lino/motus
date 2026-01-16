package trainings

import (
	"fmt"
	"strings"

	"github.com/gi8lino/motus/internal/utils"
)

// NewStateFromWorkout builds a TrainingState for the SPA from the stored workout definition.
func NewStateFromWorkout(workout *Workout, soundURLByKey func(string) string) TrainingState {
	state := TrainingState{
		TrainingID:   utils.NewID(),
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
				pauseState := TrainingStepState{
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
					stepState := TrainingStepState{
						ID:                     subsetBase,
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
					stepState := TrainingStepState{
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
				restState := TrainingStepState{
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
