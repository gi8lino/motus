import type { TrainingState } from "../types";

export function getTotalElapsedMillis(
  training: TrainingState | null,
  currentElapsedMillis: number,
): number {
  if (!training) return 0;
  return training.steps.reduce(
    (total, step, index) =>
      total +
      (index === training.currentIndex
        ? currentElapsedMillis
        : step.elapsedMillis || 0),
    0,
  );
}
