import type { TrainingStepState } from "../../types";

export function getStepSoundTargetSeconds(
  step: TrainingStepState,
  exerciseDurationSeconds: number,
): number {
  if (step.type === "pause") return step.estimatedSeconds || 0;
  if (step.superset || step.exercises?.length !== 1) return 0;
  return step.estimatedSeconds || exerciseDurationSeconds;
}

export function didStepChange(
  previousStepKey: string | null,
  currentStepKey: string,
): boolean {
  return previousStepKey !== null && previousStepKey !== currentStepKey;
}
