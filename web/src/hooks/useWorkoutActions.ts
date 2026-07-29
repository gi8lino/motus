import { useCallback } from "react";
import { createWorkout, updateWorkout } from "../api";
import type { AskConfirmOptions, Workout } from "../types";
import { MESSAGES, PROMPTS, toErrorMessage } from "../utils/messages";
import { duplicateWorkoutDraft } from "../components/workouts/workoutDraftReducer";

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
  notify: (message: string) => Promise<void>;
  currentUserId: string | null;

  /**
   * Optional persistence hooks.
   * If you already have API helpers (delete/share endpoints), plug them in here
   * without changing the UI components.
   */
  deleteWorkoutApi?: (workoutId: string) => Promise<void>;
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
  notify,
  currentUserId,
  deleteWorkoutApi,
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

  const duplicateWorkout = useCallback(
    async (workoutId: string) => {
      const workout = workouts.find((w) => w.id === workoutId);
      if (!workout || !currentUserId) {
        await notify(PROMPTS.workoutNotFound);
        return;
      }
      try {
        const created = await createWorkout({
          userId: currentUserId,
          ...duplicateWorkoutDraft(workout),
        });
        setWorkouts((current) => (current ? [created, ...current] : [created]));
        await notify("Workout duplicated.");
      } catch (err) {
        await notify(toErrorMessage(err, MESSAGES.saveWorkoutFailed));
      }
    },
    [currentUserId, notify, setWorkouts, workouts],
  );

  const updateOrganization = useCallback(
    async (workoutId: string, patch: Pick<Workout, "favorite" | "tags">) => {
      const workout = workouts.find((item) => item.id === workoutId);
      if (!workout || !currentUserId) return;
      try {
        const updated = await updateWorkout(workoutId, {
          ...workout,
          ...patch,
          userId: currentUserId,
        });
        setWorkouts(
          (current) =>
            current?.map((item) =>
              item.id === workoutId ? updated : item,
            ) ?? [updated],
        );
      } catch (err) {
        await notify(toErrorMessage(err, MESSAGES.updateWorkoutFailed));
      }
    },
    [currentUserId, notify, setWorkouts, workouts],
  );

  return {
    newWorkout,
    editWorkoutFromList,
    removeWorkout,
    duplicateWorkout,
    updateOrganization,
  };
}
