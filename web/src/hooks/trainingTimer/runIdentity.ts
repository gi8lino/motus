import type { TrainingState, TrainingStepState } from "../../types";

export function buildStepRunKey(state: TrainingState, step: TrainingStepState, estimatedSeconds: number) {
  return `${state.trainingId}:${state.currentIndex}:${step.id || ""}:${state.runningSince || 0}:${estimatedSeconds}`;
}

export function isSameStepRun(current: TrainingState, expected: TrainingState, currentStep: TrainingStepState, expectedStep: TrainingStepState, estimatedSeconds: number) {
  return current.trainingId === expected.trainingId &&
    current.currentIndex === expected.currentIndex &&
    (current.runningSince || 0) === (expected.runningSince || 0) &&
    (currentStep.id || "") === (expectedStep.id || "") &&
    (currentStep.estimatedSeconds || 0) === estimatedSeconds;
}
