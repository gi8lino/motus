import type { MutableRefObject } from "react";
import type { TrainingState, TrainingStepState } from "../../types";

// OverrunState tracks whether the modal is visible and its deadline.
export type OverrunState = {
  show: boolean;
  deadlineMs: number;
};

// OverrunRefState stores per-step overrun threshold and postpone state.
export type OverrunRefState = {
  key: string | null;
  thresholdMs: number;
  postponedUntilMs: number | null;
  hasShown: boolean;
};

// TrainingRefs bundles refs needed by overrun logic.
export type TrainingRefs = {
  trainingRef: MutableRefObject<TrainingState | null>;
  currentStepRef: MutableRefObject<TrainingStepState | null>;
  elapsedRef: MutableRefObject<number>;
};

// UseTrainingOverrunArgs configures the overrun hook.
export type UseTrainingOverrunArgs = {
  training: TrainingState | null;
  currentStep: TrainingStepState | null;
  elapsed: number;
  onPause: () => void;
  refs: TrainingRefs;
};
