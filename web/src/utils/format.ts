import type { Exercise, WorkoutStep } from "../types";

// Format milliseconds into m:ss for clocks and labels.
export function formatMillis(ms: number) {
  const totalSeconds = Math.max(0, Math.floor(ms / 1000));
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

// formatExerciseLine renders an exercise as "amount × name (weight)".
export function formatExerciseLine(ex: Exercise) {
  const amount = (ex.amount || "").trim();
  const name = (ex.name || "").trim();
  const weight = (ex.weight || "").trim();
  const cleanWeight = weight === "__auto__" ? "" : weight;
  let base = "";
  if (amount && name) {
    base = `${amount} × ${name}`;
  } else {
    base = name || amount;
  }
  if (!base) return "";
  if (cleanWeight) {
    return `${base} (${cleanWeight})`;
  }
  return base;
}

// Render exercises of a step as a compact string.
export function formatExercises(
  step:
    | WorkoutStep
    | (WorkoutStep & { elapsedMillis?: number; exercises?: any }),
) {
  // Pause steps never show exercise details.
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
    // Filter empty entries so the UI doesn't show blank pills.
    .filter((ex: Exercise) => ex && (ex.name || ex.amount || ex.weight))
    .map((ex: Exercise) => formatExerciseLine(ex))
    .filter(Boolean);
  return parts.join(" | ");
}
