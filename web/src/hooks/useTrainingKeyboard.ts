import { useEffect } from "react";
import { createTrainingKeyboardHandler } from "./trainingKeyboard/handler";
import type { UseTrainingKeyboardArgs } from "./trainingKeyboard/types";

// useTrainingKeyboard wires keyboard shortcuts for training control.
export function useTrainingKeyboard(args: UseTrainingKeyboardArgs) {
  useEffect(() => {
    const handler = createTrainingKeyboardHandler(args);
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [
    args.handleOverrunPostpone,
    args.handleOverrunPause,
    args.handlePause,
    args.handleStart,
    args.stopActiveAudio,
    args.onNext,
    args.onFinishTraining,
    args.setFinishSummary,
    args.overrunModalRef,
    args.trainingRef,
  ]);
}
