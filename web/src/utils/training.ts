import type { Exercise, TrainingState, TrainingStepState } from "../types";
import { STEP_TYPE_PAUSE } from "./step";
import { UI_TEXT } from "./uiText";

type StepLike = Partial<TrainingStepState> | null | undefined;

// getExercises normalizes the exercises list for a step payload.
export function getExercises(step: StepLike): Exercise[] {
  if (!step) return [];
  if (Array.isArray(step.exercises)) return step.exercises;
  return [];
}

// getStepName resolves the display label for a step.
export function getStepName(step: StepLike): string {
  const subsetLabel = String(step?.subsetLabel || "").trim();
  if (subsetLabel) return subsetLabel;

  const parent = String(step?.setName || "").trim();
  if (parent) return parent;

  const name = String(step?.name || "").trim();
  if (name) return name;

  if (step?.type === STEP_TYPE_PAUSE) return UI_TEXT.labels.pause;
  return UI_TEXT.labels.step;
}

// formatStepCounter renders "Step X/Y" for the current training position.
export function formatStepCounter(current: number, total: number): string {
  return `${UI_TEXT.labels.step} ${current}/${total}`;
}

// getTrainingHeaderStatus renders the status line for the training header.
export function getTrainingHeaderStatus(
  training: TrainingState | null,
): string | null {
  if (!training) return null;

  const total = training.steps?.length || 0;
  const current =
    typeof training.currentIndex === "number" ? training.currentIndex + 1 : 0;

  if (training.done) return `Finished • ${total} steps`;
  if (!training.startedAt) return `Ready • ${total} steps`;
  if (training.running) return `Running • step ${current}/${total}`;
  return `Paused • step ${current}/${total}`;
}
