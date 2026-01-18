import type { Exercise } from "../types";
import { isDurationExercise } from "./exercise";

// Shared helper: formats whole minutes/seconds as MM:SS
function formatMMSS(totalSeconds: number): string {
  const safe = Math.max(0, totalSeconds);
  const minutes = Math.floor(safe / 60);
  const seconds = safe % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

// Shared helper: formats whole hours/minutes/seconds as HH:MM:SS
function formatHHMMSS(totalSeconds: number): string {
  const safe = Math.max(0, totalSeconds);
  const hours = Math.floor(safe / 3600);
  const minutes = Math.floor((safe % 3600) / 60);
  const seconds = safe % 60;
  return `${String(hours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

// Elapsed time (stopwatch semantics)
// 19.9s → 00:19
export function formatElapsedMillis(
  ms: number,
  options?: { showHours?: boolean },
): string {
  if (!Number.isFinite(ms) || ms <= 0)
    return options?.showHours ? "00:00:00" : "00:00";
  const totalSeconds = Math.floor(ms / 1000);
  return options?.showHours
    ? formatHHMMSS(totalSeconds)
    : formatMMSS(totalSeconds);
}

// Remaining time (countdown semantics)
// 19.1s → 00:20
export function formatCountdownMillis(
  ms: number,
  options?: { showHours?: boolean },
): string {
  if (!Number.isFinite(ms) || ms <= 0)
    return options?.showHours ? "00:00:00" : "00:00";
  const totalSeconds = Math.ceil(ms / 1000);
  return options?.showHours
    ? formatHHMMSS(totalSeconds)
    : formatMMSS(totalSeconds);
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
