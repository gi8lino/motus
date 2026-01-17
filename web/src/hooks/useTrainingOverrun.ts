import { useCallback, useEffect, useRef, useState } from "react";
import type { MutableRefObject } from "react";

import type { TrainingState, TrainingStepState } from "../types";

type OverrunState = {
  show: boolean;
  deadlineMs: number;
};

type OverrunRefState = {
  key: string | null;
  thresholdMs: number;
  postponedUntilMs: number | null;
  hasShown: boolean;
};

type TrainingRefs = {
  trainingRef: MutableRefObject<TrainingState | null>;
  currentStepRef: MutableRefObject<TrainingStepState | null>;
  elapsedRef: MutableRefObject<number>;
};

type UseTrainingOverrunArgs = {
  training: TrainingState | null;
  currentStep: TrainingStepState | null;
  elapsed: number;
  onPause: () => void;
  refs: TrainingRefs;
};

const OVERRUN_GRACE_MS = 30_000;
const OVERRUN_MODAL_MS = 60_000;
const OVERRUN_COUNTDOWN_TICK_MS = 250;

// useTrainingOverrun manages the "over target" modal countdown logic.
export function useTrainingOverrun({
  training,
  currentStep,
  elapsed,
  onPause,
  refs,
}: UseTrainingOverrunArgs) {
  const { trainingRef, elapsedRef } = refs;

  const [overrunModal, setOverrunModal] = useState<OverrunState | null>(null);
  const [overrunCountdown, setOverrunCountdown] = useState(0);

  const overrunModalRef = useRef<OverrunState | null>(overrunModal);
  useEffect(() => {
    overrunModalRef.current = overrunModal;
  }, [overrunModal]);

  const overrunTimeoutRef = useRef<number | null>(null);
  const overrunIntervalRef = useRef<number | null>(null);

  const overrunRef = useRef<OverrunRefState>({
    key: null,
    thresholdMs: 0,
    postponedUntilMs: null,
    hasShown: false,
  });

  const clearOverrunTimers = useCallback(() => {
    if (overrunTimeoutRef.current) {
      clearTimeout(overrunTimeoutRef.current);
      overrunTimeoutRef.current = null;
    }
    if (overrunIntervalRef.current) {
      clearInterval(overrunIntervalRef.current);
      overrunIntervalRef.current = null;
    }
  }, []);

  const resetOverrunState = useCallback(() => {
    clearOverrunTimers();
    setOverrunModal(null);
    setOverrunCountdown(0);
  }, [clearOverrunTimers]);

  const handleOverrunPause = useCallback(() => {
    onPause();
    resetOverrunState();
  }, [onPause, resetOverrunState]);

  const handleOverrunPostpone = useCallback(() => {
    const s = trainingRef.current;
    if (!s?.running) return;

    overrunRef.current.postponedUntilMs = elapsedRef.current + OVERRUN_GRACE_MS;
    overrunRef.current.hasShown = false;
    resetOverrunState();
  }, [elapsedRef, resetOverrunState, trainingRef]);

  useEffect(() => () => resetOverrunState(), [resetOverrunState]);

  useEffect(() => {
    if (!training?.running) resetOverrunState();
  }, [training?.running, resetOverrunState]);

  useEffect(() => {
    resetOverrunState();

    const estimateMs = currentStep?.estimatedSeconds
      ? currentStep.estimatedSeconds * 1000
      : 0;

    const key =
      training?.trainingId && typeof training.currentIndex === "number"
        ? `${training.trainingId}:${training.currentIndex}:${currentStep?.id || ""}`
        : null;

    overrunRef.current.key = key;
    overrunRef.current.thresholdMs =
      estimateMs > 0 ? estimateMs + OVERRUN_GRACE_MS : 0;
    overrunRef.current.postponedUntilMs = null;
    overrunRef.current.hasShown = false;
  }, [
    training?.trainingId,
    training?.currentIndex,
    currentStep?.id,
    currentStep?.estimatedSeconds,
    resetOverrunState,
  ]);

  useEffect(() => {
    if (!training?.running || !currentStep?.estimatedSeconds) return;
    if (overrunModal?.show) return;

    const thresholdMs = overrunRef.current.thresholdMs;
    if (thresholdMs <= 0) return;

    const postponedUntilMs = overrunRef.current.postponedUntilMs;
    const effectiveThreshold =
      postponedUntilMs && postponedUntilMs > thresholdMs
        ? postponedUntilMs
        : thresholdMs;

    if (elapsed < effectiveThreshold) return;
    if (overrunRef.current.hasShown) return;

    overrunRef.current.hasShown = true;

    const deadlineMs = Date.now() + OVERRUN_MODAL_MS;
    setOverrunModal({ show: true, deadlineMs });
    setOverrunCountdown(OVERRUN_MODAL_MS);

    clearOverrunTimers();
    overrunTimeoutRef.current = window.setTimeout(() => {
      handleOverrunPause();
    }, OVERRUN_MODAL_MS);

    overrunIntervalRef.current = window.setInterval(() => {
      setOverrunCountdown(Math.max(0, deadlineMs - Date.now()));
    }, OVERRUN_COUNTDOWN_TICK_MS);
  }, [
    training?.running,
    currentStep?.estimatedSeconds,
    elapsed,
    overrunModal?.show,
    clearOverrunTimers,
    handleOverrunPause,
  ]);

  useEffect(() => {
    if (!overrunModal?.show) {
      clearOverrunTimers();
      setOverrunCountdown(0);
      return;
    }
    setOverrunCountdown(Math.max(0, overrunModal.deadlineMs - Date.now()));
    return () => clearOverrunTimers();
  }, [overrunModal?.show, overrunModal?.deadlineMs, clearOverrunTimers]);

  return {
    overrunModal,
    overrunCountdown,
    overrunModalRef,
    handleOverrunPause,
    handleOverrunPostpone,
  };
}
