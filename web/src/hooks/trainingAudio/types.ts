import type { MutableRefObject } from "react";
import type { SoundOption, TrainingState, TrainingStepState } from "../../types";

// TrainingRefs bundles mutable refs used by audio scheduling logic.
export type TrainingRefs = {
  trainingRef: MutableRefObject<TrainingState | null>;
  currentStepRef: MutableRefObject<TrainingStepState | null>;
  elapsedRef: MutableRefObject<number>;
};

// SoundSchedule stores the currently scheduled target-sound metadata.
export type SoundSchedule = {
  key: string | null;
  triggerAtMs: number;
  soundUrl: string;
};

// UseTrainingAudioArgs configures audio behavior for the training view.
export type UseTrainingAudioArgs = {
  training: TrainingState | null;
  currentStep: TrainingStepState | null;
  sounds: SoundOption[];
  markSoundPlayed: () => void;
  onPause: () => void;
  onStartStep: () => void;
  onToast: (message: string) => void;
  pauseOnTabHidden: boolean;
  refs: TrainingRefs;
};
