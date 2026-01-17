export const COUNTDOWN_GRACE_MS = 700;

// getCountdownDisplayMillis returns the value used for display (never negative).
export function getCountdownDisplayMillis(
  durationMs: number,
  elapsedMs: number,
): number {
  const remaining = Math.max(0, durationMs - Math.max(0, elapsedMs));
  return remaining;
}

// getCountdownAutoAdvanceDelay returns the delay until auto-advance should fire.
export function getCountdownAutoAdvanceDelay(
  durationMs: number,
  elapsedMs: number,
): number {
  const targetMs = durationMs + COUNTDOWN_GRACE_MS;
  return Math.max(0, targetMs - Math.max(0, elapsedMs));
}
