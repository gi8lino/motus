import { useEffect } from "react";
import type { MutableRefObject } from "react";

import type { TrainingState } from "../types";

type UseTrainingKeyboardArgs = {
  trainingRef: MutableRefObject<TrainingState | null>;
  overrunModalRef: MutableRefObject<{ show: boolean } | null>;
  handleOverrunPostpone: () => void;
  handleOverrunPause: () => void;
  handlePause: () => void;
  handleStart: () => void;
  stopActiveAudio: () => void;
  onNext: () => void;
  onFinishTraining: () => Promise<string | null>;
  setFinishSummary: (summary: string | null) => void;
};

// useTrainingKeyboard wires keyboard shortcuts for training control.
export function useTrainingKeyboard({
  trainingRef,
  overrunModalRef,
  handleOverrunPostpone,
  handleOverrunPause,
  handlePause,
  handleStart,
  stopActiveAudio,
  onNext,
  onFinishTraining,
  setFinishSummary,
}: UseTrainingKeyboardArgs) {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.repeat) return;

      const overrun = overrunModalRef.current;
      const s = trainingRef.current;

      if (overrun?.show) {
        if (e.code === "Enter") {
          e.preventDefault();
          handleOverrunPostpone();
        } else if (e.code === "Space") {
          e.preventDefault();
          handleOverrunPause();
        }
        return;
      }

      const tag = (e.target as HTMLElement | null)?.tagName;
      if (tag === "INPUT" || tag === "SELECT" || tag === "TEXTAREA") return;

      if (e.code === "Space") {
        e.preventDefault();
        if (!s) return;

        if (s.running) handlePause();
        else handleStart();
        return;
      }

      if (e.code === "Enter") {
        e.preventDefault();
        if (!s || s.done || !s.startedAt) return;

        const isLast =
          s.currentIndex >= (s.steps?.length ? s.steps.length - 1 : 0);

        stopActiveAudio();

        if (!isLast) {
          onNext();
          return;
        }

        onFinishTraining().then((summary) => {
          if (summary) setFinishSummary(summary);
        });
      }
    };

    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [
    handleOverrunPostpone,
    handleOverrunPause,
    handlePause,
    handleStart,
    stopActiveAudio,
    onNext,
    onFinishTraining,
    setFinishSummary,
    overrunModalRef,
    trainingRef,
  ]);
}
