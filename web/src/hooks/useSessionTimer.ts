import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { logSessionCompletion } from "../api";
import type { SessionState, SessionStepState } from "../types";
import { normalizeTimestamp, parseDurationSeconds } from "../utils/time";

// STORAGE_KEY stores the persisted session payload.
const STORAGE_KEY = "motus:session";

// UseSessionTimerArgs configures the session timer hook.
type UseSessionTimerArgs = {
  currentUserId?: string | null;
  onChange?: (state: SessionState | null) => void;
};

// NormalizedState adds bookkeeping metadata to session state.
type NormalizedState = SessionState & { lastUpdatedAt: number };

// now returns the current timestamp in milliseconds.
function now() {
  return Date.now();
}

// normalizeSession sanitizes stored session data into a consistent shape.
function normalizeSession(raw: SessionState): NormalizedState {
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
  const currentIndex =
    typeof raw.currentIndex === "number" ? raw.currentIndex : 0;
  base.currentIndex = Math.min(
    Math.max(currentIndex, 0),
    Math.max(raw.steps.length - 1, 0),
  );
  raw.steps.forEach((step, idx) => {
    const normalized: SessionStepState = {
      ...step,
      type: step.type === "pause" ? "pause" : "set",
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

// expandExerciseSteps expands set exercises into per-exercise steps with timing.
function expandExerciseSteps(state: SessionState): SessionState {
  const expanded: SessionState = { ...state, steps: [] };
  state.steps.forEach((step) => {
    if (step.type !== "set" || !step.exercises?.length) {
      expanded.steps.push(step);
      return;
    }
    step.exercises.forEach((ex, idx) => {
      const kind = ex.type === "timed" ? "timed" : "rep";
      const durSec = kind === "timed" ? parseDurationSeconds(ex.duration) : 0;
      const baseName = ex.name || step.name || `Exercise ${idx + 1}`;
      const stepSound = step.soundKey;
      const usesStepTarget =
        kind === "rep" && step.exercises?.length === 1 && step.estimatedSeconds;
      expanded.steps.push({
        ...step,
        id: `${step.id || "step"}-ex-${idx}`,
        name: baseName,
        estimatedSeconds:
          durSec || (usesStepTarget ? step.estimatedSeconds : 0),
        exercises: [ex],
        soundKey: stepSound,
        autoAdvance: kind === "timed" && durSec > 0,
      });
    });
  });
  if (!expanded.steps.length) {
    expanded.steps = state.steps;
  }
  return expanded;
}

// persistSession stores the current session state in localStorage.
function persistSession(state: NormalizedState | null) {
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

// loadPersistedSession restores the last session state from localStorage.
function loadPersistedSession(): NormalizedState | null {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw);
    if (!parsed?.sessionId) return null;
    const state = normalizeSession(parsed);
    const delta = parsed.lastUpdatedAt
      ? Math.min(now() - parsed.lastUpdatedAt, 30000)
      : 0;
    if (state.running && delta > 0) {
      const current = state.steps[state.currentIndex];
      if (current) {
        current.elapsedMillis += delta;
      }
    }
    if (state.running) {
      state.running = false;
      state.runningSince = null;
      const current = state.steps[state.currentIndex];
      if (current) current.running = false;
    }
    state.lastUpdatedAt = now();
    return state;
  } catch (err) {
    console.warn("unable to load session", err);
    localStorage.removeItem(STORAGE_KEY);
    return null;
  }
}

// clearPersistedSession removes stored session data.
function clearPersistedSession() {
  localStorage.removeItem(STORAGE_KEY);
}

// clone creates a deep copy of the session state.
function clone(state: NormalizedState): NormalizedState {
  return JSON.parse(JSON.stringify(state));
}

// accumulateElapsed adds elapsed time since the last update to the active step.
function accumulateElapsed(state: NormalizedState): NormalizedState {
  if (!state.running) return state;
  const current = state.steps[state.currentIndex];
  if (!current) return state;
  const nowTs = now();
  const delta = state.lastUpdatedAt ? nowTs - state.lastUpdatedAt : 0;
  if (delta > 0) {
    current.elapsedMillis += delta;
  }
  state.lastUpdatedAt = nowTs;
  return state;
}

// ensureStartedAt records the session start timestamp when missing.
function ensureStartedAt(state: NormalizedState) {
  if (!state.startedAt) {
    state.startedAt = new Date().toISOString();
  }
}

// useSessionTimer manages a workout session clock and persistence.
export function useSessionTimer({
  currentUserId,
  onChange,
}: UseSessionTimerArgs) {
  const initialSession = loadPersistedSession();
  const [restoredSessionId, setRestoredSessionId] = useState<string | null>(
    initialSession?.sessionId || null,
  );
  const [session, setSession] = useState<NormalizedState | null>(
    () => initialSession,
  );
  const rafRef = useRef<number | null>(null);

  useEffect(() => {
    persistSession(session);
    if (!session) {
      setRestoredSessionId(null);
    }
    onChange?.(session);
  }, [session, onChange]);

  // update applies a mutable update to the session state.
  const update = useCallback(
    (mutator: (next: NormalizedState) => NormalizedState | null) => {
      setSession((prev) => {
        if (!prev) return prev;
        const working = clone(prev);
        accumulateElapsed(working);
        const next = mutator(working);
        if (!next) return prev;
        return { ...next, lastUpdatedAt: now() };
      });
    },
    [],
  );

  // startFromState initializes the session from server state.
  const startFromState = useCallback(
    (raw: SessionState) => {
      const expanded = expandExerciseSteps(raw);
      const normalized = normalizeSession(expanded);
      if (!normalized.userId && currentUserId) {
        normalized.userId = currentUserId;
      }
      normalized.lastUpdatedAt = now();
      setSession(normalized);
      return normalized;
    },
    [currentUserId],
  );

  // startCurrentStep begins or resumes the current step.
  const startCurrentStep = useCallback(() => {
    update((next) => {
      next.running = true;
      next.runningSince = now();
      ensureStartedAt(next);
      next.steps = next.steps.map((step, idx) => ({
        ...step,
        current: idx === next.currentIndex,
        running: idx === next.currentIndex,
        completed: idx < next.currentIndex || step.completed,
      }));
      return next;
    });
  }, [currentUserId, update]);

  // pause stops the timer without completing the step.
  const pause = useCallback(() => {
    update((next) => {
      const current = next.steps[next.currentIndex];
      if (current && next.running) {
        const delta = next.lastUpdatedAt ? now() - next.lastUpdatedAt : 0;
        if (delta > 0) {
          current.elapsedMillis += delta;
        }
      }
      next.running = false;
      next.runningSince = null;
      if (current) {
        current.running = false;
      }
      return next;
    });
  }, [update]);

  // nextStep completes the current step and advances to the next one.
  const nextStep = useCallback(() => {
    update((next) => {
      const currentIdx = next.currentIndex;
      const current = next.steps[currentIdx];
      if (current) {
        const delta = next.lastUpdatedAt ? now() - next.lastUpdatedAt : 0;
        if (delta > 0) {
          current.elapsedMillis += delta;
        }
        current.completed = true;
        current.running = false;
        current.current = false;
      }
      const nextIdx = currentIdx + 1;
      if (nextIdx >= next.steps.length) {
        ensureStartedAt(next);
        next.done = true;
        next.running = false;
        next.currentIndex = Math.max(next.steps.length - 1, 0);
        next.completedAt = new Date().toISOString();
        next.runningSince = null;
        next.steps = next.steps.map((step) => ({
          ...step,
          completed: true,
          running: false,
          current: false,
        }));
      } else {
        next.currentIndex = nextIdx;
        next.running = true;
        next.runningSince = now();
        ensureStartedAt(next);
        next.steps = next.steps.map((step, idx) => ({
          ...step,
          completed: idx < nextIdx,
          current: idx === nextIdx,
          running: idx === nextIdx,
        }));
      }
      return next;
    });
  }, [update]);

  // finishAndLog completes the session and sends it to the backend.
  const finishAndLog = useCallback(async () => {
    const current = session || loadPersistedSession();
    if (!current) return { ok: false, error: "no session" };

    const next = clone(current);
    accumulateElapsed(next);
    ensureStartedAt(next);
    if (!next.startedAt) {
      next.startedAt = new Date().toISOString();
    }
    next.done = true;
    next.running = false;
    next.runningSince = null;
    next.currentIndex = Math.max(next.steps.length - 1, 0);
    const totalElapsedMs = next.steps.reduce(
      (sum, step) => sum + (step.elapsedMillis || 0),
      0,
    );
    const startMs = (() => {
      const parsed = Date.parse(next.startedAt || "");
      if (Number.isNaN(parsed)) {
        return Date.now() - totalElapsedMs;
      }
      return parsed;
    })();
    const completedMs = Math.max(startMs + totalElapsedMs, startMs + 1000);
    next.startedAt = new Date(startMs).toISOString();
    next.completedAt = new Date(completedMs).toISOString();
    next.steps = next.steps.map((step) => ({
      ...step,
      completed: true,
      running: false,
      current: false,
    }));
    next.lastUpdatedAt = now();
    const currentStep = next.steps[next.currentIndex];
    if (
      (currentStep?.type === "pause" &&
        currentStep.pauseOptions?.autoAdvance) ||
      currentStep?.autoAdvance
    ) {
      next.currentIndex = Math.min(
        next.currentIndex + 1,
        next.steps.length - 1,
      );
    }
    setSession(next);

    try {
      await logSessionCompletion({
        sessionId: next.sessionId,
        workoutId: next.workoutId,
        workoutName: next.workoutName,
        userId: next.userId || currentUserId || "",
        startedAt: next.startedAt || new Date().toISOString(),
        completedAt: next.completedAt || new Date().toISOString(),
        steps: next.steps.map((s, idx) => ({
          id: s.id || `step-${idx}`,
          name: s.name,
          type: s.type,
          estimatedSeconds: s.estimatedSeconds,
          elapsedMillis: s.elapsedMillis,
        })),
      });
      setSession((prev) =>
        prev && prev.sessionId === next.sessionId
          ? { ...prev, logged: true }
          : prev,
      );
      return { ok: true, session: next };
    } catch (err: any) {
      console.warn("log session failed", err);
      return { ok: false, error: err?.message || "log failed" };
    }
  }, [currentUserId, session]);

  // markSoundPlayed flags the current step sound as played.
  const markSoundPlayed = useCallback(() => {
    update((next) => {
      const current = next.steps[next.currentIndex];
      if (current) {
        current.soundPlayed = true;
      }
      return next;
    });
  }, [update]);

  useEffect(() => {
    if (!session?.running) {
      if (rafRef.current) {
        cancelAnimationFrame(rafRef.current);
        rafRef.current = null;
      }
      return;
    }
    // tick updates elapsed time on each animation frame.
    const tick = () => {
      setSession((prev) => {
        if (!prev || !prev.running) return prev;
        const working = clone(prev);
        accumulateElapsed(working);
        return working;
      });
      rafRef.current = requestAnimationFrame(tick);
    };
    rafRef.current = requestAnimationFrame(tick);
    return () => {
      if (rafRef.current) cancelAnimationFrame(rafRef.current);
      rafRef.current = null;
    };
  }, [session?.running]);

  useEffect(() => {
    // handlePageHide persists state when leaving the page.
    const handlePageHide = () => {
      setSession((prev) => {
        if (!prev?.running) return prev;
        const next = clone(prev);
        accumulateElapsed(next);
        next.running = false;
        next.runningSince = null;
        if (next.steps[next.currentIndex]) {
          next.steps[next.currentIndex].running = false;
        }
        persistSession(next);
        return next;
      });
    };
    window.addEventListener("pagehide", handlePageHide);
    window.addEventListener("beforeunload", handlePageHide);
    return () => {
      window.removeEventListener("pagehide", handlePageHide);
      window.removeEventListener("beforeunload", handlePageHide);
    };
  }, [session?.running]);

  useEffect(() => {
    // logCompletion submits a finished session if needed.
    const logCompletion = async () => {
      if (!session || !session.done || session.logged) return;
      if (!session.startedAt || !session.completedAt) {
        return;
      }
      try {
        await logSessionCompletion({
          sessionId: session.sessionId,
          workoutId: session.workoutId,
          workoutName: session.workoutName,
          userId: session.userId || currentUserId || "",
          startedAt: session.startedAt,
          completedAt: session.completedAt,
        });
        setSession((prev) =>
          prev && prev.sessionId === session.sessionId
            ? { ...prev, logged: true }
            : prev,
        );
      } catch (err) {
        console.warn("log session failed", err);
      }
    };
    logCompletion();
  }, [session, currentUserId]);

  const currentStep = useMemo(() => {
    if (!session || !session.steps.length) return null;
    return session.steps[session.currentIndex] || null;
  }, [session]);

  const displayedElapsed = useMemo(() => {
    if (!session || !currentStep) return 0;
    if (!session.running) return currentStep.elapsedMillis || 0;
    const delta = session.lastUpdatedAt ? now() - session.lastUpdatedAt : 0;
    return (currentStep.elapsedMillis || 0) + Math.max(delta, 0);
  }, [session, currentStep]);

  return {
    session,
    currentStep,
    displayedElapsed,
    restoredFromStorage: Boolean(restoredSessionId),
    startFromState,
    startCurrentStep,
    pause,
    nextStep,
    finishAndLog,
    markSoundPlayed,
    clear: () => {
      setSession(null);
      setRestoredSessionId(null);
      clearPersistedSession();
    },
  };
}
