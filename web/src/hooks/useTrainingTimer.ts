import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { logTrainingCompletion } from "../api";
import type { TrainingState } from "../types";
import { getCountdownAutoAdvanceDelay } from "../utils/countdown";
import { MESSAGES, toErrorMessage } from "../utils/messages";
import { logTimerEvent } from "../utils/timerLogger";
import { now, structuredCloneSafe } from "./trainingTimer/clock";
import { expandExerciseSteps } from "./trainingTimer/expansion";
import {
  addRunningDeltaToCurrentStep,
  applyStepFlags,
  advanceIndex,
  completeTraining,
  currentStepElapsedNow,
  isAutoAdvanceStep,
  normalizeTraining,
  setRunning,
} from "./trainingTimer/state";
import { clearPersistedTraining, loadPersistedTraining, persistTraining } from "./trainingTimer/storage";
import type { NormalizedState } from "./trainingTimer/types";

// UseTrainingTimerArgs configures the train timer hook.
type UseTrainingTimerArgs = {
  currentUserId?: string | null;
  onChange?: (state: TrainingState | null) => void;
};

// useTrainingTimer owns training progression, persistence, and auto-advance timing.
export function useTrainingTimer({
  currentUserId,
  onChange,
}: UseTrainingTimerArgs) {
  const initialTraining = loadPersistedTraining();

  const [restoredTrainingId, setRestoredTrainingId] = useState<string | null>(
    initialTraining?.trainingId || null,
  );
  const [training, setTraining] = useState<NormalizedState | null>(
    () => initialTraining,
  );

  // Render clock driven by RAF while running.
  const [nowMs, setNowMs] = useState(() => now());
  const rafIdRef = useRef<number | null>(null);

  const trainingRef = useRef<NormalizedState | null>(initialTraining);

  // Auto-advance scheduler state.
  const autoAdvanceRef = useRef<{
    timeoutId: number | null;
    key: string | null;
  }>({ timeoutId: null, key: null });

  // Stable step run anchor for deadline computation.
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
    persistTraining(training);
    if (!training) setRestoredTrainingId(null);
    onChange?.(training);
  }, [training, onChange]);

  // update applies a mutable update with consistent elapsed accumulation.
  const update = useCallback(
    (mutator: (next: NormalizedState) => NormalizedState | null) => {
      setTraining((prev) => {
        if (!prev) return prev;

        const at = now();
        const working = structuredCloneSafe(prev);

        addRunningDeltaToCurrentStep(working, at);

        const next = mutator(working);
        if (!next) return prev;

        next.lastUpdatedAt = at;
        return next;
      });
    },
    [],
  );

  // startFromState initializes training from server state.
  const startFromState = useCallback(
    (raw: TrainingState) => {
      const expanded = expandExerciseSteps(raw);
      const normalized = normalizeTraining(expanded);

      if (!normalized.userId && currentUserId) {
        normalized.userId = currentUserId;
      }

      normalized.lastUpdatedAt = now();
      applyStepFlags(normalized);

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
    const current = trainingRef.current;
    if (current) {
      const at = now();
      logTimerEvent("pause-step", {
        trainingId: current.trainingId,
        currentIndex: current.currentIndex ?? 0,
        stepId: current.steps?.[current.currentIndex ?? 0]?.id,
        elapsedMs: currentStepElapsedNow(current, at),
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
      const current = trainingRef.current;
      if (current) {
        const at = now();
        const step = current.steps?.[current.currentIndex ?? 0];
        const payload = {
          trainingId: current.trainingId,
          currentIndex: current.currentIndex ?? 0,
          stepId: step?.id || step?.name,
          elapsedMs: currentStepElapsedNow(current, at),
        };

        if (reason === "auto") {
          logTimerEvent("auto-advance-step", { ...payload, triggered: true });
        } else {
          logTimerEvent("advance-step", payload);
        }
      }

      update((next) => {
        next.running = false;
        next.runningSince = null;
        applyStepFlags(next);

        stepRunRef.current = { key: null, startedAtMs: 0 };

        advanceIndex(next);
        return next;
      });
    },
    [update],
  );

  // finishAndLog completes training and sends it to the backend.
  const finishAndLog = useCallback(async () => {
    const current = trainingRef.current || loadPersistedTraining();
    if (!current) return { ok: false, error: "no train" };

    if (current.logged) return { ok: true, training: current };
    if (current.done) return { ok: true, training: current };

    if (finishingRef.current === current.trainingId) {
      return { ok: true };
    }

    finishingRef.current = current.trainingId;

    try {
      const at = now();
      const next = structuredCloneSafe(current);

      addRunningDeltaToCurrentStep(next, at);
      completeTraining(next);

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
          steps: next.steps.map((step, idx) => ({
            id: step.id || `step-${idx}`,
            name: step.name,
            type: step.type,
            estimatedSeconds: step.estimatedSeconds,
            elapsedMillis: step.elapsedMillis,
          })),
        });

        setTraining((prev) =>
          prev && prev.trainingId === next.trainingId
            ? { ...prev, logged: true }
            : prev,
        );

        return { ok: true, training: next };
      } catch (err) {
        console.warn("log train failed", err);
        return { ok: false, error: toErrorMessage(err, MESSAGES.logTrainingFailed) };
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

  // RAF render loop while running.
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

    setNowMs(now());
    rafIdRef.current = requestAnimationFrame(loop);

    return () => {
      cancelled = true;
      stop();
    };
  }, [training?.running]);

  // Auto-advance timed steps.
  useEffect(() => {
    const clear = () => {
      if (autoAdvanceRef.current.timeoutId) {
        clearTimeout(autoAdvanceRef.current.timeoutId);
      }
      autoAdvanceRef.current.timeoutId = null;
      autoAdvanceRef.current.key = null;
    };

    const state = training;
    if (!state || !state.running || state.done) {
      clear();
      return;
    }

    const step = state.steps?.[state.currentIndex];
    if (!step) {
      clear();
      return;
    }

    const estimatedSeconds = step.estimatedSeconds || 0;
    if (!isAutoAdvanceStep(step) || estimatedSeconds <= 0) {
      clear();
      return;
    }

    const runKey = `${state.trainingId}:${state.currentIndex}:${step.id || ""}:${state.runningSince || 0}:${estimatedSeconds}`;

    if (stepRunRef.current.key !== runKey) {
      const at = now();
      const elapsedAt = currentStepElapsedNow(state, at);
      stepRunRef.current = {
        key: runKey,
        startedAtMs: at - Math.max(0, elapsedAt),
      };
    }

    const durationMs = estimatedSeconds * 1000;
    const at = now();
    const elapsedAt = Math.max(0, at - stepRunRef.current.startedAtMs);
    const remainingMs = getCountdownAutoAdvanceDelay(durationMs, elapsedAt);

    if (autoAdvanceRef.current.key === runKey) return;

    if (autoAdvanceRef.current.timeoutId) {
      clearTimeout(autoAdvanceRef.current.timeoutId);
    }
    autoAdvanceRef.current.key = runKey;

    const fire = () => {
      const current = trainingRef.current;
      if (!current || !current.running || current.done) return;

      const currentStep = current.steps?.[current.currentIndex ?? 0];
      if (!currentStep) return;

      const stillSameRun =
        current.trainingId === state.trainingId &&
        current.currentIndex === state.currentIndex &&
        (current.runningSince || 0) === (state.runningSince || 0) &&
        (currentStep.id || "") === (step.id || "") &&
        isAutoAdvanceStep(currentStep) &&
        (currentStep.estimatedSeconds || 0) === estimatedSeconds;

      if (!stillSameRun) return;

      const nowAtFire = now();
      const elapsedAtFire = currentStepElapsedNow(current, nowAtFire);
      const currentDurationMs = (currentStep.estimatedSeconds || 0) * 1000;
      const remaining = getCountdownAutoAdvanceDelay(
        currentDurationMs,
        elapsedAtFire,
      );

      if (currentDurationMs > 0 && remaining > 0) {
        autoAdvanceRef.current.timeoutId = window.setTimeout(fire, remaining);
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
    training?.steps?.[training?.currentIndex ?? 0]?.autoAdvance,
    training?.steps?.[training?.currentIndex ?? 0]?.pauseOptions?.autoAdvance,
    nextStep,
  ]);

  // Persist state on page hide/unload.
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

        persistTraining(next);
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

  // Submit a finished training if needed.
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
