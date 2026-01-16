import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { logSessionCompletion } from "../api";
import type { TrainState, TrainStepState } from "../types";
import { normalizeTimestamp, parseDurationSeconds } from "../utils/time";
import {
  STEP_TYPE_PAUSE,
  STEP_TYPE_SET,
  normalizeStepType,
} from "../utils/step";
import {
  EXERCISE_TYPE_COUNTDOWN,
  EXERCISE_TYPE_STOPWATCH,
  normalizeExerciseType,
} from "../utils/exercise";
import { logTimerEvent } from "../utils/timerLogger";

// STORAGE_KEY stores the persisted session payload.
const STORAGE_KEY = "motus:session";

// UseSessionTimerArgs configures the session timer hook.
type UseSessionTimerArgs = {
  currentUserId?: string | null;
  onChange?: (state: TrainState | null) => void;
};

// NormalizedState adds bookkeeping metadata to session state.
type NormalizedState = TrainState & { lastUpdatedAt: number };

// now returns the current timestamp in milliseconds.
function now() {
  return Date.now();
}

function structuredCloneSafe<T>(value: T): T {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const sc = (globalThis as any).structuredClone as
    | undefined
    | ((v: any) => any);
  if (typeof sc === "function") return sc(value) as T;
  return JSON.parse(JSON.stringify(value)) as T;
}

// isAutoAdvanceStep returns true when a step should auto-advance at timer end.
function isAutoAdvanceStep(step: TrainStepState | null | undefined): boolean {
  if (!step) return false;
  if (step.type === STEP_TYPE_PAUSE) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return Boolean((step as any).pauseOptions?.autoAdvance);
  }
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return Boolean((step as any).autoAdvance);
}

// normalizeSession sanitizes stored session data into a consistent shape.
function normalizeSession(raw: TrainState): NormalizedState {
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
    const normalized: TrainStepState = {
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

// expandExerciseSteps expands set exercises into per-exercise steps with timing.
function expandExerciseSteps(state: TrainState): TrainState {
  const expanded: TrainState = { ...state, steps: [] };
  const sourceSteps = Array.isArray(state.steps) ? state.steps : [];

  sourceSteps.forEach((step) => {
    const shouldExpand =
      step.type === STEP_TYPE_SET &&
      (step.exercises?.length || 0) > 1 &&
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      !Boolean((step as any).superset);

    if (!shouldExpand) {
      expanded.steps.push(step);
      return;
    }

    step.exercises?.forEach((ex, idx) => {
      const kind = normalizeExerciseType(ex.type);
      const durSec = parseDurationSeconds(ex.duration);
      const baseName = ex.name || step.name || `Exercise ${idx + 1}`;
      const stepSound = step.soundKey;

      const usesStepTarget =
        kind === "rep" &&
        step.exercises?.length === 1 &&
        Boolean(step.estimatedSeconds);

      const isDurationExercise =
        kind === EXERCISE_TYPE_STOPWATCH || kind === EXERCISE_TYPE_COUNTDOWN;

      expanded.steps.push({
        ...step,
        id: `${step.id || "step"}-ex-${idx}`,
        name: baseName,
        estimatedSeconds: isDurationExercise
          ? durSec
          : usesStepTarget
            ? step.estimatedSeconds
            : undefined,
        exercises: [ex],
        soundKey: stepSound,
        autoAdvance: kind === EXERCISE_TYPE_COUNTDOWN && durSec > 0,
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
    console.warn("unable to load session", err);
    localStorage.removeItem(STORAGE_KEY);
    return null;
  }
}

// clearPersistedSession removes stored session data.
function clearPersistedSession() {
  localStorage.removeItem(STORAGE_KEY);
}

// ensureStartedAt records the session start timestamp when missing.
function ensureStartedAt(state: NormalizedState) {
  if (!state.startedAt) state.startedAt = new Date().toISOString();
}

// applyStepFlags normalizes step flags based on currentIndex/running/done.
function applyStepFlags(state: NormalizedState) {
  const idx = state.currentIndex ?? 0;
  state.steps = (state.steps || []).map((step, i) => {
    const completed = state.done ? true : Boolean(step.completed || i < idx);
    const current = !state.done && i === idx;
    const running = Boolean(state.running) && current;
    return { ...step, completed, current, running };
  });
}

// addRunningDeltaToCurrentStep accumulates elapsedMillis based on lastUpdatedAt.
function addRunningDeltaToCurrentStep(state: NormalizedState, atMs: number) {
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

// currentStepElapsedNow reads elapsed for the active step *right now*
// without mutating state.
function currentStepElapsedNow(state: NormalizedState, atMs: number): number {
  const step = state.steps?.[state.currentIndex];
  if (!step) return 0;
  if (!state.running) return step.elapsedMillis || 0;

  const last = state.lastUpdatedAt || atMs;
  const delta = Math.max(0, atMs - last);
  return (step.elapsedMillis || 0) + delta;
}

// setRunning toggles running state and stamps runningSince.
function setRunning(state: NormalizedState, running: boolean) {
  state.running = running && !state.done;
  state.runningSince = state.running ? now() : null;
  ensureStartedAt(state);
  applyStepFlags(state);
}

// advanceIndex moves to next step, handling superset skip + done condition.
function advanceIndex(state: NormalizedState) {
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

// completeSession stamps done state + stable timestamps based on elapsed sum.
function completeSession(state: NormalizedState) {
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

  // Render clock at 10Hz without mutating session state constantly.
  const [nowMs, setNowMs] = useState(() => now());

  const sessionRef = useRef<NormalizedState | null>(initialSession);
  const tickTimerRef = useRef<number | null>(null);

  // Auto-advance scheduler.
  const autoAdvanceRef = useRef<{
    timeoutId: number | null;
    key: string | null;
  }>({
    timeoutId: null,
    key: null,
  });

  // Stable step run "start time" in wall-clock ms for accurate deadlines.
  const stepRunRef = useRef<{ key: string | null; startedAtMs: number }>({
    key: null,
    startedAtMs: 0,
  });

  // Finish/log concurrency guard.
  const finishingRef = useRef<string | null>(null);

  useEffect(() => {
    sessionRef.current = session;
  }, [session]);

  // Persist + notify parent.
  useEffect(() => {
    persistSession(session);
    if (!session) setRestoredSessionId(null);
    onChange?.(session);
  }, [session, onChange]);

  // update applies a mutable update to the session state, with consistent elapsed accumulation.
  const update = useCallback(
    (mutator: (next: NormalizedState) => NormalizedState | null) => {
      setSession((prev) => {
        if (!prev) return prev;

        const at = now();
        const working = structuredCloneSafe(prev);

        // Accumulate elapsed once per state transition.
        addRunningDeltaToCurrentStep(working, at);

        const next = mutator(working);
        if (!next) return prev;

        next.lastUpdatedAt = at;
        return next;
      });
    },
    [],
  );

  // startFromState initializes the session from server state.
  const startFromState = useCallback(
    (raw: TrainState) => {
      const expanded = expandExerciseSteps(raw);
      const normalized = normalizeSession(expanded);

      if (!normalized.userId && currentUserId) {
        normalized.userId = currentUserId;
      }

      normalized.lastUpdatedAt = now();
      applyStepFlags(normalized);

      // Reset step run wall-clock anchor.
      stepRunRef.current = { key: null, startedAtMs: 0 };

      setSession(normalized);
      return normalized;
    },
    [currentUserId],
  );

  // startCurrentStep begins or resumes the current step.
  const startCurrentStep = useCallback(() => {
    update((next) => {
      setRunning(next, true);
      return next;
    });
  }, [update]);

  // pause stops the timer without completing the step.
  const pause = useCallback(() => {
    const cur = sessionRef.current;
    if (cur) {
      const at = now();
      logTimerEvent("pause-step", {
        sessionId: cur.sessionId,
        currentIndex: cur.currentIndex ?? 0,
        stepId: cur.steps?.[cur.currentIndex ?? 0]?.id,
        elapsedMs: currentStepElapsedNow(cur, at),
      });
    }

    update((next) => {
      setRunning(next, false);
      return next;
    });
  }, [update]);

  // nextStep completes the current step and advances to the next one.
  const nextStep = useCallback(
    (reason: "manual" | "auto" = "manual") => {
      const cur = sessionRef.current;
      if (cur) {
        const at = now();
        const step = cur.steps?.[cur.currentIndex ?? 0];
        const payload = {
          sessionId: cur.sessionId,
          currentIndex: cur.currentIndex ?? 0,
          stepId: step?.id || step?.name,
          elapsedMs: currentStepElapsedNow(cur, at),
        };

        if (reason === "auto") {
          logTimerEvent("auto-advance-step", { ...payload, triggered: true });
        } else {
          logTimerEvent("advance-step", payload);
        }
      }

      update((next) => {
        // Stop running while we transition, then advance.
        next.running = false;
        next.runningSince = null;
        applyStepFlags(next);

        // Reset step run anchor; next step will re-anchor when running.
        stepRunRef.current = { key: null, startedAtMs: 0 };

        advanceIndex(next);
        return next;
      });
    },
    [update],
  );

  // finishAndLog completes the session and sends it to the backend.
  const finishAndLog = useCallback(async () => {
    const cur = sessionRef.current || loadPersistedSession();
    if (!cur) return { ok: false, error: "no session" };

    if (finishingRef.current === cur.sessionId) {
      return { ok: false, error: "already finishing" };
    }
    finishingRef.current = cur.sessionId;

    try {
      const at = now();
      const next = structuredCloneSafe(cur);

      // Finalize elapsed if running.
      addRunningDeltaToCurrentStep(next, at);

      completeSession(next);

      // Preserve your prior behavior: if last step is auto-advance, nudge index.
      const last = next.steps?.[next.currentIndex];
      if (isAutoAdvanceStep(last)) {
        next.currentIndex = Math.min(
          next.currentIndex + 1,
          next.steps.length - 1,
        );
      }

      setSession(next);

      logTimerEvent("finish-session", {
        sessionId: next.sessionId,
        workoutId: next.workoutId,
        currentIndex: next.currentIndex ?? 0,
        steps: next.steps?.length || 0,
      });

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
    } finally {
      finishingRef.current = null;
    }
  }, [currentUserId]);

  // markSoundPlayed flags the current step sound as played.
  const markSoundPlayed = useCallback(() => {
    update((next) => {
      const step = next.steps?.[next.currentIndex];
      if (step) step.soundPlayed = true;
      return next;
    });
  }, [update]);

  // Drive clock rendering at 10Hz while running.
  useEffect(() => {
    if (!session?.running) {
      if (tickTimerRef.current) {
        window.clearInterval(tickTimerRef.current);
        tickTimerRef.current = null;
      }
      return;
    }

    tickTimerRef.current = window.setInterval(() => {
      setNowMs(now());
    }, 100);

    return () => {
      if (tickTimerRef.current) {
        window.clearInterval(tickTimerRef.current);
        tickTimerRef.current = null;
      }
    };
  }, [session?.running]);

  // Auto-advance timed steps:
  // - schedules once per run instance
  // - uses stable stepRunRef for "deadline computation"
  // - logs only when it actually triggers
  useEffect(() => {
    const clear = () => {
      if (autoAdvanceRef.current.timeoutId) {
        clearTimeout(autoAdvanceRef.current.timeoutId);
      }
      autoAdvanceRef.current.timeoutId = null;
      autoAdvanceRef.current.key = null;
    };

    const s = session;
    if (!s || !s.running || s.done) {
      clear();
      return;
    }

    const step = s.steps?.[s.currentIndex];
    if (!step) {
      clear();
      return;
    }

    const estimatedSeconds = step.estimatedSeconds || 0;
    if (!isAutoAdvanceStep(step) || estimatedSeconds <= 0) {
      clear();
      return;
    }

    // Build a stable key for this "run instance"
    // (pause/resume changes runningSince; step change changes currentIndex/id)
    const runKey = `${s.sessionId}:${s.currentIndex}:${step.id || ""}:${s.runningSince || 0}:${estimatedSeconds}`;

    // Anchor a wall-clock start for this run, once.
    if (stepRunRef.current.key !== runKey) {
      const at = now();
      const elapsedAt = currentStepElapsedNow(s, at);
      stepRunRef.current = {
        key: runKey,
        startedAtMs: at - Math.max(0, elapsedAt),
      };
    }

    // Compute deadline from anchored start.
    const durationMs = estimatedSeconds * 1000;
    const deadlineMs = stepRunRef.current.startedAtMs + durationMs;

    const at = now();
    const remainingMs = Math.max(0, deadlineMs - at);

    // Don’t reschedule if already scheduled for this runKey.
    if (autoAdvanceRef.current.key === runKey) {
      return;
    }

    // Replace existing schedule.
    if (autoAdvanceRef.current.timeoutId) {
      clearTimeout(autoAdvanceRef.current.timeoutId);
    }
    autoAdvanceRef.current.key = runKey;

    const fire = () => {
      const cur = sessionRef.current;
      if (!cur || !cur.running || cur.done) return;

      const curStep = cur.steps?.[cur.currentIndex ?? 0];
      if (!curStep) return;

      const stillSameRun =
        cur.sessionId === s.sessionId &&
        cur.currentIndex === s.currentIndex &&
        (cur.runningSince || 0) === (s.runningSince || 0) &&
        (curStep.id || "") === (step.id || "") &&
        isAutoAdvanceStep(curStep) &&
        (curStep.estimatedSeconds || 0) === estimatedSeconds;

      if (!stillSameRun) return;

      // Guard: if timers were clamped/delayed, ensure the step is actually finished.
      const at2 = now();
      const elapsedAtFire = currentStepElapsedNow(cur, at2);
      const durMs = (curStep.estimatedSeconds || 0) * 1000;

      if (durMs > 0 && elapsedAtFire < durMs) {
        const rem = Math.max(0, durMs - elapsedAtFire);
        autoAdvanceRef.current.timeoutId = window.setTimeout(fire, rem);
        return;
      }

      nextStep("auto");
    };

    if (remainingMs <= 0) {
      fire();
      return;
    }

    autoAdvanceRef.current.timeoutId = window.setTimeout(fire, remainingMs);
    return clear;
  }, [
    session?.sessionId,
    session?.currentIndex,
    session?.running,
    session?.runningSince,
    session?.done,
    // only depend on fields that affect auto-advance behavior
    session?.steps?.[session?.currentIndex ?? 0]?.id,
    session?.steps?.[session?.currentIndex ?? 0]?.type,
    session?.steps?.[session?.currentIndex ?? 0]?.estimatedSeconds,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (session?.steps?.[session?.currentIndex ?? 0] as any)?.autoAdvance,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (session?.steps?.[session?.currentIndex ?? 0] as any)?.pauseOptions
      ?.autoAdvance,
    nextStep,
  ]);

  // Persist state on page hide/unload (finalize elapsed, then stop).
  useEffect(() => {
    const handlePageHide = () => {
      setSession((prev) => {
        if (!prev) return prev;

        const at = now();
        const next = structuredCloneSafe(prev);

        addRunningDeltaToCurrentStep(next, at);

        next.running = false;
        next.runningSince = null;
        next.lastUpdatedAt = at;
        applyStepFlags(next);

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
  }, []);

  // Submit a finished session if needed (no extra timer logs).
  useEffect(() => {
    const logCompletion = async () => {
      if (!session || !session.done || session.logged) return;
      if (!session.startedAt || !session.completedAt) return;

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
    if (!session || !session.steps?.length) return null;
    return session.steps[session.currentIndex] || null;
  }, [session]);

  const displayedElapsed = useMemo(() => {
    if (!session || !currentStep) return 0;

    // While running, compute elapsed relative to nowMs (10Hz tick).
    if (!session.running) return currentStep.elapsedMillis || 0;

    const at = nowMs;
    return currentStepElapsedNow(session, at);
  }, [session, currentStep, nowMs]);

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
