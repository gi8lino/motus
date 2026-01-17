import type { TrainingState, TrainingStepState } from "../../types";
import { normalizeTimestamp } from "../../utils/time";
import { STEP_TYPE_PAUSE, normalizeStepType } from "../../utils/step";
import { now } from "./time";

// NormalizedState adds bookkeeping metadata to train state.
export type NormalizedState = TrainingState & { lastUpdatedAt: number };

// isAutoAdvanceStep returns true when a step should auto-advance at timer end.
export function isAutoAdvanceStep(
  step: TrainingStepState | null | undefined,
): boolean {
  if (!step) return false;
  if (step.type === STEP_TYPE_PAUSE) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return Boolean((step as any).pauseOptions?.autoAdvance);
  }
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return Boolean((step as any).autoAdvance);
}

// normalizeTraining sanitizes stored train data into a consistent shape.
export function normalizeTraining(raw: TrainingState): NormalizedState {
  const base: NormalizedState = {
    ...raw,
    running: Boolean(raw.running && !raw.done),
    runningSince: raw.runningSince || null,
    done: Boolean(raw.done),
    startedAt: normalizeTimestamp(raw.startedAt),
    completedAt: normalizeTimestamp(raw.completedAt),
    logged: Boolean(raw.logged),
    lastUpdatedAt: now(),
    steps: [],
  };

  const rawSteps = Array.isArray(raw.steps) ? raw.steps : [];
  const currentIndex =
    typeof raw.currentIndex === "number" ? raw.currentIndex : 0;

  base.currentIndex = Math.min(
    Math.max(currentIndex, 0),
    Math.max(rawSteps.length - 1, 0),
  );

  rawSteps.forEach((step, idx) => {
    const normalized: TrainingStepState = {
      ...step,
      type: normalizeStepType(step.type),
      elapsedMillis: step.elapsedMillis || 0,
      completed: Boolean(step.completed || idx < base.currentIndex),
      current: idx === base.currentIndex && !base.done,
      running: Boolean(step.running) && idx === base.currentIndex && !base.done,
      soundPlayed: Boolean(step.soundPlayed),
    };
    base.steps.push(normalized);
  });

  if (base.done) {
    base.running = false;
    base.runningSince = null;
    base.steps = base.steps.map((step) => ({
      ...step,
      running: false,
      current: false,
      completed: true,
    }));
  }

  return base;
}

// ensureStartedAt records the train start timestamp when missing.
export function ensureStartedAt(state: NormalizedState) {
  if (!state.startedAt) state.startedAt = new Date().toISOString();
}

// applyStepFlags normalizes step flags based on currentIndex/running/done.
export function applyStepFlags(state: NormalizedState) {
  const idx = state.currentIndex ?? 0;
  state.steps = (state.steps || []).map((step, i) => {
    const completed = state.done ? true : Boolean(step.completed || i < idx);
    const current = !state.done && i === idx;
    const running = Boolean(state.running) && current;
    return { ...step, completed, current, running };
  });
}

// addRunningDeltaToCurrentStep accumulates elapsedMillis based on lastUpdatedAt.
export function addRunningDeltaToCurrentStep(
  state: NormalizedState,
  atMs: number,
) {
  if (!state.running) {
    state.lastUpdatedAt = atMs;
    return;
  }
  const step = state.steps?.[state.currentIndex];
  if (!step) {
    state.lastUpdatedAt = atMs;
    return;
  }
  const last = state.lastUpdatedAt || atMs;
  const delta = Math.max(0, atMs - last);
  if (delta > 0) {
    step.elapsedMillis = (step.elapsedMillis || 0) + delta;
  }
  state.lastUpdatedAt = atMs;
}

// currentStepElapsedNow reads elapsed for the active step *right now* without mutating state.
export function currentStepElapsedNow(
  state: NormalizedState,
  atMs: number,
): number {
  const step = state.steps?.[state.currentIndex];
  if (!step) return 0;
  if (!state.running) return step.elapsedMillis || 0;

  const last = state.lastUpdatedAt || atMs;
  const delta = Math.max(0, atMs - last);
  return (step.elapsedMillis || 0) + delta;
}

// setRunning toggles running state and stamps runningSince.
export function setRunning(state: NormalizedState, running: boolean) {
  state.running = running && !state.done;
  state.runningSince = state.running ? now() : null;
  ensureStartedAt(state);
  applyStepFlags(state);
}

// advanceIndex moves to next step, handling superset skip + done condition.
export function advanceIndex(state: NormalizedState) {
  const currentIdx = state.currentIndex ?? 0;
  const current = state.steps?.[currentIdx];

  // Mark current completed.
  if (current) {
    current.completed = true;
    current.current = false;
    current.running = false;
  }

  // Superset skip behavior (preserving your original logic).
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const skipSubsetId =
    current && Boolean((current as any).superset) && (current as any).subsetId
      ? // eslint-disable-next-line @typescript-eslint/no-explicit-any
        String((current as any).subsetId)
      : null;

  let nextIdx = currentIdx + 1;
  if (skipSubsetId) {
    while (nextIdx < (state.steps?.length || 0)) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const candidate = state.steps[nextIdx] as any;
      if (candidate?.subsetId === skipSubsetId) {
        nextIdx += 1;
        continue;
      }
      break;
    }
  }

  if (!state.steps?.length || nextIdx >= state.steps.length) {
    // Done.
    state.done = true;
    state.running = false;
    state.runningSince = null;
    state.currentIndex = Math.max((state.steps?.length || 1) - 1, 0);
    state.completedAt = new Date().toISOString();
    applyStepFlags(state);
    return;
  }

  // Move to next.
  state.currentIndex = nextIdx;
  state.running = true;
  state.runningSince = now();
  ensureStartedAt(state);
  applyStepFlags(state);
}

// completeTraining stamps done state + stable timestamps based on elapsed sum.
export function completeTraining(state: NormalizedState) {
  ensureStartedAt(state);
  state.done = true;
  state.running = false;
  state.runningSince = null;
  state.currentIndex = Math.max((state.steps?.length || 1) - 1, 0);

  const totalElapsedMs = (state.steps || []).reduce(
    (sum, step) => sum + (step.elapsedMillis || 0),
    0,
  );

  const startedAtMs = (() => {
    const parsed = Date.parse(state.startedAt || "");
    if (!Number.isFinite(parsed)) return now() - totalElapsedMs;
    return parsed;
  })();

  const completedAtMs = Math.max(
    startedAtMs + totalElapsedMs,
    startedAtMs + 1000,
  );

  state.startedAt = new Date(startedAtMs).toISOString();
  state.completedAt = new Date(completedAtMs).toISOString();

  state.steps = (state.steps || []).map((step) => ({
    ...step,
    completed: true,
    running: false,
    current: false,
  }));
}
