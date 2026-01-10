import type { Exercise, WorkoutStep } from "../types";
import { isDurationExercise } from "./exercise";
import { STEP_TYPE_PAUSE } from "./step";

// Format milliseconds into m:ss for clocks and labels.
export function formatMillis(ms: number) {
  const totalSeconds = Math.max(0, Math.floor(ms / 1000));
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

// formatExerciseLine renders an exercise based on its type.
export function formatExerciseLine(ex: Exercise) {
  const kind = ex.type || "rep";
  const reps = (ex.reps || "").trim();
  const name = (ex.name || "").trim();
  const weight = (ex.weight || "").trim();
  const duration = (ex.duration || "").trim();
  if (isDurationExercise(kind)) {
    const displayName = name || "";
    if (!displayName && !duration) return "";
    let base = displayName || duration;
    if (displayName && duration) {
      base = `${displayName} ${duration}`;
    }
    return base;
  }
  let base = "";
  if (reps && name) {
    base = `${reps} × ${name}`;
  } else {
    base = name || reps;
  }
  if (!base) return "";
  if (weight) {
    return `${base} (${weight})`;
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
  if (step.type === STEP_TYPE_PAUSE) return "";
  const list = step.exercises || [];
  const parts = list
    // Filter empty entries so the UI doesn't show blank pills.
    .filter(
      (ex: Exercise) => ex && (ex.name || ex.reps || ex.weight || ex.duration),
    )
    .map((ex: Exercise) => formatExerciseLine(ex))
    .filter(Boolean);
  return parts.join(" | ");
}
