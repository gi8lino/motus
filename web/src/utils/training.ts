import type { Exercise, TrainingState, TrainingStepState } from "../types";
import { formatCountdownMillis, formatExerciseLine } from "./format";
import { STEP_TYPE_PAUSE } from "./step";
import { UI_TEXT } from "./uiText";

type AnyStep = any;

export type StepGroup = {
  key: string;
  setName: string;
  type: string;
  loopIndex: number;
  loopTotal: number;
  current: boolean;
  subsets: Array<{
    key: string;
    superset: boolean;
    label?: string;
    exercises: Exercise[];
  }>;
  hasSuperset: boolean;
  estimatedSeconds: number;
};

// getExercises normalizes the exercises list for a step payload.
export function getExercises(step: AnyStep): Exercise[] {
  if (!step) return [];
  if (Array.isArray(step.exercises)) return step.exercises;
  return [];
}

// getStepName resolves the display label for a step.
export function getStepName(step: AnyStep): string {
  const subsetLabel = String(step?.subsetLabel || "").trim();
  if (subsetLabel) return subsetLabel;

  const parent = String(step?.setName || "").trim();
  if (parent) return parent;

  const name = String(step?.name || "").trim();
  if (name) return name;

  if (step?.type === STEP_TYPE_PAUSE) return UI_TEXT.labels.pause;
  return UI_TEXT.labels.step;
}

// getCurrentExerciseLabel builds the display label for the active exercise.
export function getCurrentExerciseLabel(step: AnyStep): string {
  if (!step) return "";
  const exercise = Array.isArray(step.exercises)
    ? step.exercises[0]
    : undefined;
  const formatted = exercise ? formatExerciseLine(exercise) : "";
  return formatted || getStepName(step);
}

// extractExerciseLabels builds pill text for a step (superset-aware).
export function extractExerciseLabels(
  step: TrainingStepState | null,
  training?: TrainingState | null,
  startIndex = 0,
): string[] {
  if (!step) return [];

  if (step.type === STEP_TYPE_PAUSE) {
    const pauseLabel = getStepName(step);
    const durationText = step.estimatedSeconds
      ? formatCountdownMillis(step.estimatedSeconds * 1000)
      : "";
    const pauseText = durationText
      ? `${pauseLabel} • ${durationText}`
      : pauseLabel;
    return [pauseText];
  }

  if (step.superset && step.subsetId && training?.steps?.length) {
    const subsetId = step.subsetId;
    const seen = new Set<string>();
    const labels: string[] = [];
    for (let idx = startIndex; idx < training.steps.length; idx += 1) {
      const candidate = training.steps[idx];
      if (candidate.subsetId !== subsetId) continue;
      for (const ex of getExercises(candidate)) {
        const text = formatExerciseLine(ex);
        if (text && !seen.has(text)) {
          seen.add(text);
          labels.push(text);
        }
      }
    }
    if (labels.length) return labels;
  }

  return getExercises(step)
    .map((ex) => formatExerciseLine(ex))
    .filter(Boolean);
}

// getNextStep resolves the next step to show in the "Next" panel.
export function getNextStep(
  training: TrainingState | null,
): TrainingStepState | null {
  if (!training) return null;

  if (!training.running) {
    return training.steps[training.currentIndex];
  }

  let idx = training.currentIndex;
  while (idx < training.steps.length && training.steps[idx].completed) idx += 1;
  idx += 1;
  while (idx < training.steps.length && training.steps[idx].completed) idx += 1;

  return idx < training.steps.length ? training.steps[idx] : null;
}

// getNextSubsetStep finds the next step belonging to a different subset.
export function getNextSubsetStep(
  training: TrainingState | null,
  currentStep: TrainingStepState | null,
): TrainingStepState | null {
  if (!training || !training.steps.length || !currentStep?.subsetId)
    return null;

  const subsetId = currentStep.subsetId;
  const startIdx = Math.max(training.currentIndex + 1, 0);

  for (let idx = startIdx; idx < training.steps.length; idx += 1) {
    const candidate = training.steps[idx];
    if (!candidate) continue;
    if (!candidate.subsetId) continue;
    if (candidate.subsetId !== subsetId) return candidate;
  }

  return null;
}

// hasFollowingSubsetExercises reports whether the next step has more subset entries.
export function hasFollowingSubsetExercises(
  training: TrainingState | null,
  nextStep: TrainingStepState | null,
): boolean {
  if (!training || !nextStep?.subsetId) return false;
  const subsetId = nextStep.subsetId;
  return training.steps.some(
    (step, idx) => idx > training.currentIndex && step.subsetId === subsetId,
  );
}

// buildStepGroups normalizes remaining steps into display groups.
export function buildStepGroups(
  remainingSteps: TrainingStepState[],
): StepGroup[] {
  if (!remainingSteps.length) return [];

  const groups: StepGroup[] = [];
  let currentGroup: StepGroup | null = null;

  for (const step of remainingSteps) {
    const setName = String(step?.setName || step?.name || "").trim() || "Step";
    const loopIndex = step?.loopIndex ?? 0;
    const loopTotal = step?.loopTotal ?? 0;
    const type = step?.type || "set";

    const needsNewGroup =
      !currentGroup ||
      currentGroup.setName !== setName ||
      currentGroup.loopIndex !== loopIndex ||
      currentGroup.loopTotal !== loopTotal ||
      currentGroup.type !== type;

    if (needsNewGroup) {
      currentGroup = {
        key: `${setName}-${loopIndex}-${loopTotal}-${groups.length}`,
        setName,
        type,
        loopIndex,
        loopTotal,
        current: Boolean(step.current),
        subsets: [],
        hasSuperset: false,
        estimatedSeconds: step?.estimatedSeconds ?? 0,
      };
      groups.push(currentGroup);
    } else if (currentGroup) {
      currentGroup.current = currentGroup.current || Boolean(step.current);
    }

    if (!currentGroup) continue;
    if (type === STEP_TYPE_PAUSE) continue;

    const subsetKey = String(
      step?.subsetId || step?.id || `${setName}-${loopIndex}-${loopTotal}`,
    );
    let subset = currentGroup.subsets.find((item) => item.key === subsetKey);

    if (!subset) {
      subset = {
        key: subsetKey,
        superset: Boolean(step?.superset),
        label: step?.subsetLabel,
        exercises: [],
      };
      currentGroup.subsets.push(subset);
    } else {
      subset.superset = subset.superset || Boolean(step?.superset);
      if (!subset.label && step?.subsetLabel) {
        subset.label = step.subsetLabel;
      }
    }

    const exercises = getExercises(step);
    if (exercises.length) subset.exercises.push(...exercises);
    currentGroup.hasSuperset = currentGroup.hasSuperset || subset.superset;
  }

  return groups;
}

// buildExercisePills returns formatted pill labels for a group subset.
export function buildExercisePills(exercises: Exercise[]): string[] {
  return exercises.map((ex) => formatExerciseLine(ex)).filter(Boolean);
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
