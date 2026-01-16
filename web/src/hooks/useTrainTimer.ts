import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { logTrainingCompletion } from "../api";
import type { TrainngState, TrainngStepState } from "../types";
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

// STORAGE_KEY stores the persisted train payload.
const STORAGE_KEY = "motus:train";

// UseTrainingTimerArgs configures the train timer hook.
type UseTrainTimerArgs = {
  currentUserId?: string | null;
  onChange?: (state: TrainngState | null) => void;
};

// NormalizedState adds bookkeeping metadata to train state.
type NormalizedState = TrainngState & { lastUpdatedAt: number };

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
function isAutoAdvanceStep(step: TrainngStepState | null | undefined): boolean {
  if (!step) return false;
  if (step.type === STEP_TYPE_PAUSE) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return Boolean((step as any).pauseOptions?.autoAdvance);
  }
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return Boolean((step as any).autoAdvance);
}

// normalizeTrain sanitizes stored train data into a consistent shape.
function normalizeTrain(raw: TrainngState): NormalizedState {
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
    const normalized: TrainngStepState = {
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
function expandExerciseSteps(state: TrainngState): TrainngState {
  const expanded: TrainngState = { ...state, steps: [] };
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

// persistTrain stores the current train state in localStorage.
function persistTrain(state: NormalizedState | null) {
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

// loadPersistedTrain restores the last train state from localStorage.
function loadPersistedTrain(): NormalizedState | null {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;

  try {
    const parsed = JSON.parse(raw);
    if (!parsed?.trainingId) return null;

    const state = normalizeTrain(parsed);

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
function clearPersistedTraining() {
  localStorage.removeItem(STORAGE_KEY);
}

// ensureStartedAt records the train start timestamp when missing.
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

// completeTraining stamps done state + stable timestamps based on elapsed sum.
function completeTraining(state: NormalizedState) {
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

export function useTrainingTimer({
  currentUserId,
  onChange,
}: UseTrainTimerArgs) {
  const initialTraining = loadPersistedTrain();

  const [restoredTrainingId, setRestoredTrainingId] = useState<string | null>(
    initialTraining?.trainingId || null,
  );
  const [training, setTraining] = useState<NormalizedState | null>(
    () => initialTraining,
  );

  // Render clock driven by RAF while running (avoids interval clamping).
  const [nowMs, setNowMs] = useState(() => now());
  const rafIdRef = useRef<number | null>(null);

  const trainingRef = useRef<NormalizedState | null>(initialTraining);

  // Auto-advance scheduler.
  const autoAdvanceRef = useRef<{
    timeoutId: number | null;
    key: string | null;
  }>({ timeoutId: null, key: null });

  // Stable step run "start time" in wall-clock ms for accurate deadlines.
  const stepRunRef = useRef<{ key: string | null; startedAtMs: number }>({
    key: null,
    startedAtMs: 0,
  });

  // Finish/log concurrency guard.
  const finishingRef = useRef<string | null>(null);

  useEffect(() => {
    trainingRef.current = training;
  }, [training]);

  // Persist + notify parent.
  useEffect(() => {
    persistTrain(training);
    if (!training) setRestoredTrainingId(null);
    onChange?.(training);
  }, [training, onChange]);

  // update applies a mutable update to the train state, with consistent elapsed accumulation.
  const update = useCallback(
    (mutator: (next: NormalizedState) => NormalizedState | null) => {
      setTraining((prev) => {
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

  // startFromState initializes the train from server state.
  const startFromState = useCallback(
    (raw: TrainngState) => {
      const expanded = expandExerciseSteps(raw);
      const normalized = normalizeTrain(expanded);

      if (!normalized.userId && currentUserId) {
        normalized.userId = currentUserId;
      }

      normalized.lastUpdatedAt = now();
      applyStepFlags(normalized);

      // Reset step run wall-clock anchor.
      stepRunRef.current = { key: null, startedAtMs: 0 };

      setTraining(normalized);
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
    const cur = trainingRef.current;
    if (cur) {
      const at = now();
      logTimerEvent("pause-step", {
        trainingId: cur.trainingId,
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
      const cur = trainingRef.current;
      if (cur) {
        const at = now();
        const step = cur.steps?.[cur.currentIndex ?? 0];
        const payload = {
          trainingId: cur.trainingId,
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

  // finishAndLog completes the train and sends it to the backend.
  const finishAndLog = useCallback(async () => {
    const cur = trainingRef.current || loadPersistedTrain();
    if (!cur) return { ok: false, error: "no train" };

    if (finishingRef.current === cur.trainingId) {
      return { ok: false, error: "already finishing" };
    }
    finishingRef.current = cur.trainingId;

    try {
      const at = now();
      const next = structuredCloneSafe(cur);

      // Finalize elapsed if running.
      addRunningDeltaToCurrentStep(next, at);

      completeTraining(next);

      // Preserve prior behavior: if last step is auto-advance, nudge index.
      const last = next.steps?.[next.currentIndex];
      if (isAutoAdvanceStep(last)) {
        next.currentIndex = Math.min(
          next.currentIndex + 1,
          next.steps.length - 1,
        );
      }

      setTraining(next);

      logTimerEvent("finish-train", {
        trainingId: next.trainingId,
        workoutId: next.workoutId,
        currentIndex: next.currentIndex ?? 0,
        steps: next.steps?.length || 0,
      });

      try {
        await logTrainingCompletion({
          trainingId: next.trainingId,
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

        setTraining((prev) =>
          prev && prev.trainingId === next.trainingId
            ? { ...prev, logged: true }
            : prev,
        );

        return { ok: true, train: next };
      } catch (err: any) {
        console.warn("log train failed", err);
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

  // RAF render loop while running (replaces 10Hz interval).
  useEffect(() => {
    const stop = () => {
      if (rafIdRef.current != null) {
        cancelAnimationFrame(rafIdRef.current);
        rafIdRef.current = null;
      }
    };

    if (!training?.running) {
      stop();
      return;
    }

    let cancelled = false;

    const loop = () => {
      if (cancelled) return;
      setNowMs(now());
      rafIdRef.current = requestAnimationFrame(loop);
    };

    // kick immediately so the first frame isn't delayed
    setNowMs(now());
    rafIdRef.current = requestAnimationFrame(loop);

    return () => {
      cancelled = true;
      stop();
    };
  }, [training?.running]);

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

    const s = training;
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

    const runKey = `${s.trainingId}:${s.currentIndex}:${step.id || ""}:${s.runningSince || 0}:${estimatedSeconds}`;

    if (stepRunRef.current.key !== runKey) {
      const at = now();
      const elapsedAt = currentStepElapsedNow(s, at);
      stepRunRef.current = {
        key: runKey,
        startedAtMs: at - Math.max(0, elapsedAt),
      };
    }

    const durationMs = estimatedSeconds * 1000;
    const deadlineMs = stepRunRef.current.startedAtMs + durationMs;

    const at = now();
    const remainingMs = Math.max(0, deadlineMs - at);

    if (autoAdvanceRef.current.key === runKey) {
      return;
    }

    if (autoAdvanceRef.current.timeoutId) {
      clearTimeout(autoAdvanceRef.current.timeoutId);
    }
    autoAdvanceRef.current.key = runKey;

    const fire = () => {
      const cur = trainingRef.current;
      if (!cur || !cur.running || cur.done) return;

      const curStep = cur.steps?.[cur.currentIndex ?? 0];
      if (!curStep) return;

      const stillSameRun =
        cur.trainingId === s.trainingId &&
        cur.currentIndex === s.currentIndex &&
        (cur.runningSince || 0) === (s.runningSince || 0) &&
        (curStep.id || "") === (step.id || "") &&
        isAutoAdvanceStep(curStep) &&
        (curStep.estimatedSeconds || 0) === estimatedSeconds;

      if (!stillSameRun) return;

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
    training?.trainingId,
    training?.currentIndex,
    training?.running,
    training?.runningSince,
    training?.done,
    training?.steps?.[training?.currentIndex ?? 0]?.id,
    training?.steps?.[training?.currentIndex ?? 0]?.type,
    training?.steps?.[training?.currentIndex ?? 0]?.estimatedSeconds,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (training?.steps?.[training?.currentIndex ?? 0] as any)?.autoAdvance,
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (training?.steps?.[training?.currentIndex ?? 0] as any)?.pauseOptions
      ?.autoAdvance,
    nextStep,
  ]);

  // Persist state on page hide/unload (finalize elapsed, then stop).
  useEffect(() => {
    const handlePageHide = () => {
      setTraining((prev) => {
        if (!prev) return prev;

        const at = now();
        const next = structuredCloneSafe(prev);

        addRunningDeltaToCurrentStep(next, at);

        next.running = false;
        next.runningSince = null;
        next.lastUpdatedAt = at;
        applyStepFlags(next);

        persistTrain(next);
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

  // Submit a finished train if needed (no extra timer logs).
  useEffect(() => {
    const logCompletion = async () => {
      if (!training || !training.done || training.logged) return;
      if (!training.startedAt || !training.completedAt) return;

      try {
        await logTrainingCompletion({
          trainingId: training.trainingId,
          workoutId: training.workoutId,
          workoutName: training.workoutName,
          userId: training.userId || currentUserId || "",
          startedAt: training.startedAt,
          completedAt: training.completedAt,
        });

        setTraining((prev) =>
          prev && prev.trainingId === training.trainingId
            ? { ...prev, logged: true }
            : prev,
        );
      } catch (err) {
        console.warn("log train failed", err);
      }
    };
    logCompletion();
  }, [training, currentUserId]);

  const currentStep = useMemo(() => {
    if (!training || !training.steps?.length) return null;
    return training.steps[training.currentIndex] || null;
  }, [training]);

  const displayedElapsed = useMemo(() => {
    if (!training || !currentStep) return 0;
    if (!training.running) return currentStep.elapsedMillis || 0;
    return currentStepElapsedNow(training, nowMs);
  }, [training, currentStep, nowMs]);

  return {
    training,
    currentStep,
    displayedElapsed,
    restoredFromStorage: Boolean(restoredTrainingId),
    startFromState,
    startCurrentStep,
    pause,
    nextStep,
    finishAndLog,
    markSoundPlayed,
    clear: () => {
      setTraining(null);
      setRestoredTrainingId(null);
      clearPersistedTraining();
    },
  };
}
