import type { TrainingState, TrainingStepState } from "../../types";

const OVERRUN_GRACE_MS = 30_000;

// buildOverrunKey returns the stable key for one training-step instance.
export function buildOverrunKey(
  training: TrainingState | null,
  currentStep: TrainingStepState | null,
): string | null {
  if (!training?.trainingId || typeof training.currentIndex !== "number") {
    return null;
  }
  return `${training.trainingId}:${training.currentIndex}:${currentStep?.id || ""}`;
}

// getOverrunThresholdMs returns the threshold when overrun handling should start.
export function getOverrunThresholdMs(estimatedSeconds?: number): number {
  if (!estimatedSeconds || estimatedSeconds <= 0) return 0;
  return estimatedSeconds * 1000 + OVERRUN_GRACE_MS;
}

// getEffectiveThresholdMs returns threshold including an optional postpone value.
export function getEffectiveThresholdMs(
  thresholdMs: number,
  postponedUntilMs: number | null,
): number {
  if (postponedUntilMs && postponedUntilMs > thresholdMs) {
    return postponedUntilMs;
  }
  return thresholdMs;
}

// clearTimer clears a timeout/interval id and nulls the reference.
export function clearTimer(timerRef: { current: number | null }): void {
  if (timerRef.current) {
    clearTimeout(timerRef.current);
    timerRef.current = null;
  }
}
