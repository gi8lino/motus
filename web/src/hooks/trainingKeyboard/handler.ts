import type { UseTrainingKeyboardArgs } from "./types";

// isEditableTarget reports whether keyboard shortcuts should be ignored.
function isEditableTarget(target: EventTarget | null): boolean {
  const tag = (target as HTMLElement | null)?.tagName;
  return tag === "INPUT" || tag === "SELECT" || tag === "TEXTAREA";
}

// createTrainingKeyboardHandler returns the keydown handler for training shortcuts.
export function createTrainingKeyboardHandler({
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
}: UseTrainingKeyboardArgs): (event: KeyboardEvent) => void {
  return (event: KeyboardEvent) => {
    if (event.repeat) return;

    const overrun = overrunModalRef.current;
    const training = trainingRef.current;

    if (overrun?.show) {
      if (event.code === "Enter") {
        event.preventDefault();
        handleOverrunPostpone();
      } else if (event.code === "Space") {
        event.preventDefault();
        handleOverrunPause();
      }
      return;
    }

    if (isEditableTarget(event.target)) return;

    if (event.code === "Space") {
      event.preventDefault();
      if (!training) return;

      if (training.running) handlePause();
      else handleStart();
      return;
    }

    if (event.code !== "Enter") return;

    event.preventDefault();
    if (!training || training.done || !training.startedAt) return;

    const isLast =
      training.currentIndex >=
      (training.steps?.length ? training.steps.length - 1 : 0);

    stopActiveAudio();

    if (!isLast) {
      onNext();
      return;
    }

    onFinishTraining().then((summary) => {
      if (summary) setFinishSummary(summary);
    });
  };
}
