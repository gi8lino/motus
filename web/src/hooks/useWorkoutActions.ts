import { useCallback } from "react";
import type { AskConfirmOptions, Workout } from "../types";
import { MESSAGES, PROMPTS, toErrorMessage } from "../utils/messages";
import { UI_TEXT } from "../utils/uiText";

type UseWorkoutActionsArgs = {
  workouts: Workout[];

  selectedWorkoutId: string | null;

  setEditingWorkout: (workout: Workout | null) => void;
  setSelectedWorkoutId: (id: string | null) => void;

  setWorkouts: (updater: (prev: Workout[] | null) => Workout[] | null) => void;

  // Dialog helpers (match useDialog)
  askConfirm: (
    message: string,
    options?: AskConfirmOptions,
  ) => Promise<boolean>;
  askPrompt: (message: string, defaultValue?: string) => Promise<string | null>;

  notify: (message: string) => Promise<void>;

  // Optional: refresh templates list after “share”
  templatesReload?: () => void;

  /**
   * Optional persistence hooks.
   * If you already have API helpers (delete/share endpoints), plug them in here
   * without changing the UI components.
   */
  deleteWorkoutApi?: (workoutId: string) => Promise<void>;
  shareWorkoutApi?: (workoutId: string) => Promise<void>;
};

/**
 * useWorkoutActions wires up list-level actions for the workouts view:
 * - new / edit selection
 * - delete / share (optional persistence)
 *
 * IMPORTANT: This hook does NOT open/close modals or own UI state.
 * UI components decide when to show the editor.
 */
export function useWorkoutActions({
  workouts,
  selectedWorkoutId,
  setEditingWorkout,
  setSelectedWorkoutId,
  setWorkouts,
  askConfirm,
  askPrompt,
  notify,
  templatesReload,
  deleteWorkoutApi,
  shareWorkoutApi,
}: UseWorkoutActionsArgs) {
  const newWorkout = useCallback(() => {
    setSelectedWorkoutId(null);
    setEditingWorkout(null);
  }, [setEditingWorkout, setSelectedWorkoutId]);

  const editWorkoutFromList = useCallback(
    (workoutId: string) => {
      const found = workouts.find((w) => w.id === workoutId) || null;
      setSelectedWorkoutId(workoutId);
      setEditingWorkout(found);
    },
    [setEditingWorkout, setSelectedWorkoutId, workouts],
  );

  const removeWorkout = useCallback(
    async (workoutId: string) => {
      const workout = workouts.find((w) => w.id === workoutId);
      const label = workout?.name ? `“${workout.name}”` : "this workout";

      const ok = await askConfirm(`Delete ${label}?`);
      if (!ok) return;

      try {
        if (deleteWorkoutApi) {
          await deleteWorkoutApi(workoutId);
        }

        // Always update local list so UI reflects deletion immediately.
        setWorkouts((prev) =>
          prev ? prev.filter((w) => w.id !== workoutId) : prev,
        );

        // Clear selection if we deleted the selected workout.
        if (selectedWorkoutId === workoutId) {
          setSelectedWorkoutId(null);
          setEditingWorkout(null);
        }

        await notify(PROMPTS.workoutDeleted);
      } catch (err) {
        await notify(toErrorMessage(err, MESSAGES.deleteWorkoutFailed));
      }
    },
    [
      askConfirm,
      deleteWorkoutApi,
      notify,
      selectedWorkoutId,
      setEditingWorkout,
      setSelectedWorkoutId,
      setWorkouts,
      workouts,
    ],
  );

  const shareWorkout = useCallback(
    async (workoutId: string) => {
      const workout = workouts.find((w) => w.id === workoutId);
      if (!workout) {
        await notify(PROMPTS.workoutNotFound);
        return;
      }

      // If you have a real share endpoint, prefer it.
      if (shareWorkoutApi) {
        try {
          await shareWorkoutApi(workoutId);
          templatesReload?.();
          await notify(UI_TEXT.toasts.shared);
          return;
        } catch (err) {
          await notify(toErrorMessage(err, MESSAGES.shareWorkoutFailed));
          return;
        }
      }

      // Fallback: “share” by copying JSON to clipboard.
      const name = workout.name || UI_TEXT.labels.workout;
      const defaultValue = `${name} (template)`;
      const templateName = await askPrompt(
        UI_TEXT.prompts.templateName,
        defaultValue,
      );
      if (templateName === null) return;

      const payload = {
        ...workout,
        name: templateName.trim() || name,
        isTemplate: true,
      };

      try {
        await navigator.clipboard.writeText(JSON.stringify(payload, null, 2));
        templatesReload?.();
        await notify(UI_TEXT.toasts.templateCopied);
      } catch (err) {
        await notify(toErrorMessage(err, MESSAGES.copyTemplateFailed));
      }
    },
    [askPrompt, notify, shareWorkoutApi, templatesReload, workouts],
  );

  return {
    newWorkout,
    editWorkoutFromList,
    removeWorkout,
    shareWorkout,
  };
}
