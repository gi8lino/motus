import type { WorkoutStep } from "../types";
import { parseDurationSeconds } from "./time";

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
      if (step.type === "set" && step.exercises?.length) {
        step.exercises.forEach((ex) => {
          const kind = ex.type === "timed" ? "timed" : "rep";
          const baseName = ex.name || step.name;
          const usesStepTarget =
            kind === "rep" &&
            (step.exercises?.length || 0) === 1 &&
            step.estimatedSeconds;
          expanded.push({
            ...step,
            name: baseName,
            exercises: [ex],
            estimatedSeconds:
              kind === "timed"
                ? parseDurationSeconds(ex.duration)
                : usesStepTarget
                  ? step.estimatedSeconds
                  : undefined,
            loopIndex,
            loopTotal,
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
