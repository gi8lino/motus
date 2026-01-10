import type { WorkoutStep } from "../types";
import { parseDurationSeconds } from "./time";
import {
  EXERCISE_TYPE_COUNTDOWN,
  EXERCISE_TYPE_REP,
  EXERCISE_TYPE_STOPWATCH,
  normalizeExerciseType,
} from "./exercise";
import { STEP_TYPE_SET } from "./step";

// ExpandedWorkoutStep adds loop metadata for repeated steps.
export type ExpandedWorkoutStep = WorkoutStep & {
  loopIndex?: number;
  loopTotal?: number;
};

// expandWorkoutSteps expands repeat definitions into a flat list of steps.
export function expandWorkoutSteps(
  steps: WorkoutStep[],
): ExpandedWorkoutStep[] {
  const expanded: ExpandedWorkoutStep[] = [];
  steps.forEach((step) => {
    const repeatCount = Math.max(1, Math.floor(step.repeatCount || 1));
    const restSeconds = Math.max(0, Math.floor(step.repeatRestSeconds || 0));
    const restAfterLast = Boolean(step.repeatRestAfterLast);
    const restSoundKey = step.repeatRestSoundKey || "";
    const restAutoAdvance = Boolean(step.repeatRestAutoAdvance);

    for (let loop = 0; loop < repeatCount; loop += 1) {
      const loopIndex = repeatCount > 1 ? loop + 1 : undefined;
      const loopTotal = repeatCount > 1 ? repeatCount : undefined;
      if (step.type === STEP_TYPE_SET && step.exercises?.length) {
        step.exercises.forEach((ex) => {
          const kind = normalizeExerciseType(ex.type);
          const baseName = ex.name || step.name;
          const usesStepTarget =
            kind === EXERCISE_TYPE_REP &&
            (step.exercises?.length || 0) === 1 &&
            step.estimatedSeconds;
          const durationSeconds = parseDurationSeconds(ex.duration);
          const isDurationExercise =
            kind === EXERCISE_TYPE_STOPWATCH ||
            kind === EXERCISE_TYPE_COUNTDOWN;
          expanded.push({
            ...step,
            name: baseName,
            exercises: [ex],
            estimatedSeconds: isDurationExercise
              ? durationSeconds
              : usesStepTarget
                ? step.estimatedSeconds
                : undefined,
            loopIndex,
            loopTotal,
            autoAdvance: kind === EXERCISE_TYPE_COUNTDOWN,
          });
        });
      } else {
        expanded.push({
          ...step,
          loopIndex,
          loopTotal,
        });
      }
      if (restSeconds > 0 && (loop < repeatCount - 1 || restAfterLast)) {
        expanded.push({
          type: "pause",
          name: "Pause",
          estimatedSeconds: restSeconds,
          soundKey: restSoundKey,
          pauseOptions: restAutoAdvance ? { autoAdvance: true } : undefined,
          loopIndex,
          loopTotal,
        });
      }
    }
  });
  return expanded;
}
