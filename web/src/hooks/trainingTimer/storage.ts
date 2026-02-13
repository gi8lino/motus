import { now } from "./clock";
import { normalizeTraining } from "./state";
import type { NormalizedState } from "./types";

// STORAGE_KEY stores the persisted train payload.
const STORAGE_KEY = "motus:train";

type PersistedTraining = Partial<NormalizedState> & {
  trainingId?: string;
  lastUpdatedAt?: number;
};

// persistTraining stores the current train state in localStorage.
export function persistTraining(state: NormalizedState | null): void {
  if (!state || state.done) {
    localStorage.removeItem(STORAGE_KEY);
    return;
  }
  localStorage.setItem(
    STORAGE_KEY,
    JSON.stringify({
      ...state,
      lastUpdatedAt: now(),
    }),
  );
}

// loadPersistedTraining restores the last train state from localStorage.
export function loadPersistedTraining(): NormalizedState | null {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;

  try {
    const parsed = JSON.parse(raw) as PersistedTraining;
    if (!parsed.trainingId) return null;

    const state = normalizeTraining(parsed as NormalizedState);

    // Never resume "running" from storage. Restore elapsed (capped) then pause.
    const delta = parsed.lastUpdatedAt
      ? Math.min(now() - parsed.lastUpdatedAt, 30_000)
      : 0;

    if (state.running && delta > 0) {
      const current = state.steps[state.currentIndex];
      if (current) current.elapsedMillis = (current.elapsedMillis || 0) + delta;
    }

    state.running = false;
    state.runningSince = null;
    const current = state.steps[state.currentIndex];
    if (current) current.running = false;

    state.lastUpdatedAt = now();
    return state;
  } catch (err) {
    console.warn("unable to load train", err);
    localStorage.removeItem(STORAGE_KEY);
    return null;
  }
}

// clearPersistedTraining removes stored train data.
export function clearPersistedTraining(): void {
  localStorage.removeItem(STORAGE_KEY);
}
