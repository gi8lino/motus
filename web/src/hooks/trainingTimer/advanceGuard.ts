import type { TrainingState, TrainingStepState } from "../../types";

type AdvanceGuardRef = { current: string | null };

function stepTransitionKey(
  training: Pick<
    TrainingState,
    "trainingId" | "currentIndex" | "runningSince" | "startedAt"
  >,
  step: Pick<TrainingStepState, "id" | "name"> | null | undefined,
): string {
  return [
    training.trainingId,
    training.currentIndex,
    step?.id || step?.name || "",
    training.runningSince || 0,
    training.startedAt || "",
  ].join(":");
}

// claimAdvanceTransition returns true once per step/run until state moves on.
export function claimAdvanceTransition(
  guardRef: AdvanceGuardRef,
  training: Pick<
    TrainingState,
    "trainingId" | "currentIndex" | "runningSince" | "startedAt" | "steps"
  > | null,
): boolean {
  if (!training) return false;

  const key = stepTransitionKey(
    training,
    training.steps?.[training.currentIndex] || null,
  );

  if (guardRef.current === key) return false;
  guardRef.current = key;
  return true;
}

// syncAdvanceTransition clears the guard once the training has moved on.
export function syncAdvanceTransition(
  guardRef: AdvanceGuardRef,
  training: Pick<
    TrainingState,
    "trainingId" | "currentIndex" | "runningSince" | "startedAt" | "steps"
  > | null,
): void {
  if (!guardRef.current) return;
  if (!training) {
    guardRef.current = null;
    return;
  }

  const key = stepTransitionKey(
    training,
    training.steps?.[training.currentIndex] || null,
  );

  if (guardRef.current !== key) {
    guardRef.current = null;
  }
}
