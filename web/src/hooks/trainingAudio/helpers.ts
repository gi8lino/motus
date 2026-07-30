import type {
  SoundOption,
  TrainingState,
  TrainingStepState,
} from "../../types";
import { resolveMediaUrl } from "../../utils/basePath";
import { parseDurationSeconds } from "../../utils/time";
import { getStepSoundTargetSeconds } from "./transitions";

// ResolvedSoundPlan contains normalized sound/timing values for scheduling.
export type ResolvedSoundPlan = {
  subsetSoundUrl: string;
  stepSoundUrl: string;
  subsetLeadSeconds: number;
  stepLeadSeconds: number;
  subsetTargetSeconds: number;
  stepTargetSeconds: number;
};

// resolveSoundPlan normalizes target sounds and timing values for a step.
export function resolveSoundPlan(
  currentStep: TrainingStepState,
  sounds: SoundOption[],
): ResolvedSoundPlan {
  const subsetSoundKey = currentStep.soundKey || "";
  const hasExerciseTarget =
    !currentStep.superset && currentStep.exercises?.length === 1;
  const hasStepTarget = currentStep.type === "pause" || hasExerciseTarget;
  const exerciseSoundKey = hasExerciseTarget
    ? currentStep.exercises?.[0]?.soundKey || ""
    : "";

  const subsetSoundOption = subsetSoundKey
    ? sounds.find((sound) => sound.key === subsetSoundKey)
    : sounds.find((sound) => sound.file === currentStep.soundUrl);

  const subsetSoundUrl = resolveMediaUrl(
    currentStep.soundUrl || subsetSoundOption?.file || "",
  );

  const exerciseSoundOption = exerciseSoundKey
    ? sounds.find((sound) => sound.key === exerciseSoundKey)
    : undefined;

  const stepSoundUrl = hasStepTarget
    ? exerciseSoundKey && exerciseSoundOption
      ? resolveMediaUrl(exerciseSoundOption.file || "")
      : subsetSoundUrl
    : "";

  const subsetLeadSeconds = subsetSoundOption?.leadSeconds ?? 0;
  const stepLeadSeconds = exerciseSoundKey
    ? (exerciseSoundOption?.leadSeconds ?? 0)
    : subsetLeadSeconds;

  const subsetTargetSeconds = currentStep.subsetEstimatedSeconds ?? 0;
  const exerciseDurationSeconds =
    parseDurationSeconds(currentStep.duration) ||
    parseDurationSeconds(currentStep.exercises?.[0]?.duration) ||
    0;
  const stepTargetSeconds = getStepSoundTargetSeconds(
    currentStep,
    exerciseDurationSeconds,
  );

  return {
    subsetSoundUrl,
    stepSoundUrl,
    subsetLeadSeconds,
    stepLeadSeconds,
    subsetTargetSeconds,
    stepTargetSeconds,
  };
}

// getSubsetElapsedMs returns elapsed time for the current subset instance.
export function getSubsetElapsedMs(
  training: TrainingState,
  currentStep: TrainingStepState,
  currentElapsedMs: number,
): number {
  if (!currentStep.subsetId) return 0;
  const targetLoop = currentStep.loopIndex ?? 0;

  return training.steps.reduce((acc, step, idx) => {
    if (step.subsetId !== currentStep.subsetId) return acc;
    const stepLoop = step.loopIndex ?? 0;
    if (stepLoop !== targetLoop) return acc;
    if (idx < training.currentIndex) return acc + (step.elapsedMillis || 0);
    if (idx === training.currentIndex) return acc + currentElapsedMs;
    return acc;
  }, 0);
}

// clearTimer clears a window timeout and nulls its ref.
export function clearTimer(timerRef: { current: number | null }): void {
  if (timerRef.current) {
    clearTimeout(timerRef.current);
    timerRef.current = null;
  }
}
