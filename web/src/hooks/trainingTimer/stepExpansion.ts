import type { TrainingState } from "../../types";
import { parseDurationSeconds } from "../../utils/time";
import { STEP_TYPE_SET } from "../../utils/step";
import {
  EXERCISE_TYPE_COUNTDOWN,
  EXERCISE_TYPE_STOPWATCH,
  normalizeExerciseType,
} from "../../utils/exercise";

// expandExerciseSteps expands set exercises into per-exercise steps with timing.
export function expandExerciseSteps(state: TrainingState): TrainingState {
  const expanded: TrainingState = { ...state, steps: [] };
  const sourceSteps = Array.isArray(state.steps) ? state.steps : [];

  sourceSteps.forEach((step) => {
    const shouldExpand =
      step.type === STEP_TYPE_SET &&
      (step.exercises?.length || 0) > 1 &&
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      !Boolean((step as any).superset);

    if (!shouldExpand) {
      expanded.steps.push(step);
      return;
    }

    step.exercises?.forEach((ex, idx) => {
      const kind = normalizeExerciseType(ex.type);
      const durSec = parseDurationSeconds(ex.duration);
      const baseName = ex.name || step.name || `Exercise ${idx + 1}`;
      const stepSound = step.soundKey;

      const usesStepTarget =
        kind === "rep" &&
        step.exercises?.length === 1 &&
        Boolean(step.estimatedSeconds);

      const isDurationExercise =
        kind === EXERCISE_TYPE_STOPWATCH || kind === EXERCISE_TYPE_COUNTDOWN;

      expanded.steps.push({
        ...step,
        id: `${step.id || "step"}-ex-${idx}`,
        name: baseName,
        estimatedSeconds: isDurationExercise
          ? durSec
          : usesStepTarget
            ? step.estimatedSeconds
            : undefined,
        exercises: [ex],
        soundKey: stepSound,
        autoAdvance: kind === EXERCISE_TYPE_COUNTDOWN && durSec > 0,
      });
    });
  });

  if (!expanded.steps.length) {
    expanded.steps = state.steps;
  }

  return expanded;
}
