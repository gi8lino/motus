import type { MutableRefObject } from "react";
import type { TrainingState } from "../../types";

// UseTrainingKeyboardArgs configures keyboard shortcuts for training control.
export type UseTrainingKeyboardArgs = {
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
