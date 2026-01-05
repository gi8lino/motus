import type { Exercise, WorkoutStep } from "../types";

// Format milliseconds into m:ss for clocks and labels.
export function formatMillis(ms: number) {
  const totalSeconds = Math.max(0, Math.floor(ms / 1000));
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

// Render exercises of a step as a compact string.
export function formatExercises(
  step:
    | WorkoutStep
    | (WorkoutStep & { elapsedMillis?: number; exercises?: any }),
) {
  if (step.type === "pause") return "";
  const list =
    step.exercises && step.exercises.length
      ? step.exercises
      : [
          {
            name: (step as any).exercise,
            amount: (step as any).amount,
            weight: (step as any).weight,
          },
        ];
  const parts = list
    .filter((ex: Exercise) => ex && (ex.name || ex.amount || ex.weight))
    .map((ex: Exercise) => {
      const weight = ex.weight === "__auto__" ? "" : ex.weight;
      return [ex.name, ex.amount, weight].filter(Boolean).join(" • ");
    });
  return parts.join(" | ");
}
