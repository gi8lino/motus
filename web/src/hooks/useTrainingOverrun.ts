import { useCallback, useEffect, useRef, useState } from "react";

import {
  OVERRUN_COUNTDOWN_TICK_MS,
  OVERRUN_GRACE_MS,
  OVERRUN_MODAL_MS,
} from "./trainingOverrun/constants";
import {
  buildOverrunKey,
  clearTimer,
  getEffectiveThresholdMs,
  getOverrunThresholdMs,
} from "./trainingOverrun/helpers";
import type { OverrunRefState, OverrunState, UseTrainingOverrunArgs } from "./trainingOverrun/types";

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
    clearTimer(overrunTimeoutRef);
    clearTimer(overrunIntervalRef);
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
    const current = trainingRef.current;
    if (!current?.running) return;

    overrunRef.current.postponedUntilMs = elapsedRef.current + OVERRUN_GRACE_MS;
    overrunRef.current.hasShown = false;
    resetOverrunState();
  }, [elapsedRef, resetOverrunState, trainingRef]);

  useEffect(() => () => resetOverrunState(), [resetOverrunState]);

  useEffect(() => {
    if (!training?.running) resetOverrunState();
  }, [training?.running, resetOverrunState]);

  // Recompute overrun thresholds when switching to another step.
  useEffect(() => {
    resetOverrunState();

    overrunRef.current.key = buildOverrunKey(training, currentStep);
    overrunRef.current.thresholdMs = getOverrunThresholdMs(
      currentStep?.estimatedSeconds,
    );
    overrunRef.current.postponedUntilMs = null;
    overrunRef.current.hasShown = false;
  }, [
    training?.trainingId,
    training?.currentIndex,
    currentStep?.id,
    currentStep?.estimatedSeconds,
    resetOverrunState,
  ]);

  // Open modal once elapsed passes threshold.
  useEffect(() => {
    if (!training?.running || !currentStep?.estimatedSeconds) return;
    if (overrunModal?.show) return;

    const thresholdMs = overrunRef.current.thresholdMs;
    if (thresholdMs <= 0) return;

    const effectiveThresholdMs = getEffectiveThresholdMs(
      thresholdMs,
      overrunRef.current.postponedUntilMs,
    );

    if (elapsed < effectiveThresholdMs) return;
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

  // Keep countdown in sync while modal is visible.
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
