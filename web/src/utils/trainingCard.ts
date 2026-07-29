import type { Exercise, TrainingStepState } from "../types";

export function formatExerciseMetric(exercise: Exercise | undefined) {
  if (!exercise) return "";
  const parts: string[] = [];

  if (exercise.reps) parts.push(`${exercise.reps} reps`);
  else if (exercise.duration) parts.push(exercise.duration);

  if (exercise.weight) parts.push(exercise.weight);

  return parts.join(" · ");
}

export function formatExerciseSide(exercise: Exercise | undefined) {
  if (exercise?.side === "left") return "Left";
  if (exercise?.side === "right") return "Right";
  return "";
}

export function formatRoundValue(step: TrainingStepState | null) {
  const total = step?.loopTotal ?? 0;
  return total > 1 ? `${step?.loopIndex || 1}/${total}` : "";
}

export function formatStepValue(current: number, total: number) {
  return total ? `${current}/${total}` : "";
}
