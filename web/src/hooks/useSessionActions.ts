import { startSession as startSessionApi } from "../api";
import type { AskConfirmOptions, SessionState } from "../types";
import { buildSummary } from "../utils/summary";

// UseSessionActionsArgs describes dependencies for session actions.
type UseSessionActionsArgs = {
  selectedWorkoutId: string | null;
  session: SessionState | null;
  currentWorkoutName: string;
  setSessionsView: () => void;
  setPromptedResume: (next: boolean) => void;
  setResumeSuppressed: (next: boolean) => void;
  startFromState: (state: SessionState) => void;
  finishAndLog: () => Promise<{
    ok: boolean;
    error?: string;
    session?: SessionState;
  } | null>;
  historyReload: () => void;
  askConfirm: (
    message: string,
    options?: AskConfirmOptions,
  ) => Promise<boolean>;
  notify: (message: string) => Promise<void>;
};

// useSessionActions bundles start/finish actions for the active session.
export function useSessionActions({
  selectedWorkoutId,
  session,
  currentWorkoutName,
  setSessionsView,
  setPromptedResume,
  setResumeSuppressed,
  startFromState,
  finishAndLog,
  historyReload,
  askConfirm,
  notify,
}: UseSessionActionsArgs) {
  // startSession begins a new session or resumes an existing one.
  const startSession = async () => {
    // Guard: require a workout selection before starting.
    if (!selectedWorkoutId) {
      await notify("Select a workout first.");
      return;
    }
    // Guard: prompt to resume if an active session already exists.
    if (session && !session.done) {
      const resume = await askConfirm(
        "You have an active session. Resume it instead?",
        { confirmLabel: "Resume", cancelLabel: "New session" },
      );
      if (resume) {
        setSessionsView();
        return;
      }
    }
    const state = await startSessionApi(selectedWorkoutId);
    startFromState(state);
    setSessionsView();
    setPromptedResume(false);
    setResumeSuppressed(false);
  };

  // finishSession finalizes a session and returns the AI summary text.
  const finishSession = async (): Promise<string | null> => {
    const result = await finishAndLog();
    if (!result?.ok) {
      // Guard: surface API errors without generating a summary.
      await notify(result?.error || "Unable to save session");
      return null;
    }
    if (result.session) {
      historyReload();
      return buildSummary({
        workoutName: currentWorkoutName || result.session.workoutName,
        workoutId: result.session.workoutId,
        userId: result.session.userId,
        startedAt: result.session.startedAt,
        completedAt: result.session.completedAt,
        steps: result.session.steps,
      });
    }
    historyReload();
    return null;
  };

  return { startSession, finishSession };
}
