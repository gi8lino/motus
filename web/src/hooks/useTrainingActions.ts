import { startTraining as startTrainingApi } from "../api";
import type { AskConfirmOptions, TrainingState } from "../types";
import { buildSummary } from "../utils/summary";

// UseTrainingActionsArgs describes dependencies for training actions.
type UseTrainingActionsArgs = {
  selectedWorkoutId: string | null;
  training: TrainingState | null;
  currentWorkoutName: string;
  setTrainingView: () => void;
  setPromptedResume: (next: boolean) => void;
  setResumeSuppressed: (next: boolean) => void;
  startFromState: (state: TrainingState) => void;
  finishAndLog: () => Promise<{
    ok: boolean;
    error?: string;
    training?: TrainingState;
  } | null>;
  historyReload: () => void;
  askConfirm: (
    message: string,
    options?: AskConfirmOptions,
  ) => Promise<boolean>;
  notify: (message: string) => Promise<void>;
};

// useTrainingActions bundles start/finish actions for the active training.
export function useTrainingActions({
  selectedWorkoutId,
  training,
  currentWorkoutName,
  setTrainingView,
  setPromptedResume,
  setResumeSuppressed,
  startFromState,
  finishAndLog,
  historyReload,
  askConfirm,
  notify,
}: UseTrainingActionsArgs): {
  startTraining: () => Promise<void>;
  finishTraining: () => Promise<string | null>;
} {
  // startTraining prepares a new training session (or resumes if available).
  const startTraining = async () => {
    if (!selectedWorkoutId) {
      await notify("Select a workout first.");
      return;
    }

    setTrainingView();

    if (training && !training.done) {
      const resume = await askConfirm(
        "You have an active training. Resume it instead?",
        { confirmLabel: "Resume", cancelLabel: "New session" },
      );
      if (resume) {
        setPromptedResume(false);
        setResumeSuppressed(true);
        return;
      }
    }

    const state = await startTrainingApi(selectedWorkoutId);
    startFromState(state);
  };
  // finishTraining finalizes a training and returns the AI summary text.
  const finishTraining = async (): Promise<string | null> => {
    const result = await finishAndLog();
    if (!result?.ok) {
      // Guard: surface API errors without generating a summary.
      await notify(result?.error || "Unable to save training.");
      return null;
    }

    if (result.training) {
      historyReload();
      return buildSummary({
        workoutName: currentWorkoutName || result.training.workoutName,
        workoutId: result.training.workoutId,
        userId: result.training.userId,
        startedAt: result.training.startedAt,
        completedAt: result.training.completedAt,
        steps: result.training.steps,
      });
    }

    historyReload();
    return null;
  };

  return { startTraining, finishTraining };
}
