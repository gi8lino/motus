import { startTrain } from "../api";
import type { AskConfirmOptions, TrainingState } from "../types";
import { buildSummary } from "../utils/summary";

// UseTrainingActionsArgs describes dependencies for training actions.
type UseTrainingActionsArgs = {
  selectedWorkoutId: string | null;
  training: TrainingState | null;
  currentWorkoutName: string;
  setTrainView: () => void;
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

// useTrainActions bundles start/finish actions for the active training.
export function useTrainActions({
  selectedWorkoutId,
  training,
  currentWorkoutName,
  setTrainView: setTrainView,
  setPromptedResume,
  setResumeSuppressed,
  startFromState,
  finishAndLog,
  historyReload,
  askConfirm,
  notify,
}: UseTrainingActionsArgs) {
  // startTraining begins a new training or resumes an existing one.
  const startTraining = async () => {
    // Guard: require a workout selection before starting.
    if (!selectedWorkoutId) {
      await notify("Select a workout first.");
      return;
    }

    // Guard: prompt to resume if an active training already exists.
    if (training && !training.done) {
      const resume = await askConfirm(
        "You have an active workout. Resume it instead?",
        { confirmLabel: "Resume", cancelLabel: "New workout" },
      );
      if (resume) {
        setTrainView();
        return;
      }
    }

    try {
      const state = await startTrain(selectedWorkoutId);

      startFromState(state);
      setTrainView();

      // Only clear resume flags after success (so failures don’t lose state).
      setPromptedResume(false);
      setResumeSuppressed(false);
    } catch (err: any) {
      await notify(err?.message || "Unable to start training");
    }
  };

  // finishTrain finalizes a training and returns the AI summary text.
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
