import type { Exercise, TrainingStepState } from "../types";

export function formatExerciseMetric(exercise: Exercise | undefined) {
  if (!exercise) return "";
  if (exercise.reps && exercise.weight)
    return `${exercise.reps} reps · ${exercise.weight}`;
  if (exercise.reps) return `${exercise.reps} reps`;
  if (exercise.duration && exercise.weight)
    return `${exercise.duration} · ${exercise.weight}`;
  return exercise.duration || exercise.weight || "";
}

export function formatRoundValue(step: TrainingStepState | null) {
  const total = step?.loopTotal ?? 0;
  return total > 1 ? `${step?.loopIndex || 1}/${total}` : "";
}

export function formatStepValue(current: number, total: number) {
  return total ? `${current}/${total}` : "";
}
