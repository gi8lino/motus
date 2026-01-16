import {
  createWorkout,
  getWorkout,
  updateWorkout as updateWorkoutApi,
} from "../api";
import type { AskConfirmOptions, Workout, WorkoutStep } from "../types";

// UseWorkoutFormActionsArgs describes dependencies for workout form actions.
type UseWorkoutFormActionsArgs = {
  currentUserId: string | null;
  workoutDirty: boolean;
  setSelectedWorkoutId: (id: string | null) => void;
  setEditingWorkout: (workout: Workout | null) => void;
  setWorkoutDirty: (dirty: boolean) => void;
  setShowWorkoutForm: (show: boolean) => void;
  setWorkouts: (updater: (prev: Workout[] | null) => Workout[] | null) => void;
  reloadWorkouts?: () => void;
  askConfirm: (
    message: string,
    options?: AskConfirmOptions,
  ) => Promise<boolean>;
  notify?: (message: string) => Promise<void>;
};

function errorMessage(err: unknown, fallback: string) {
  return err instanceof Error ? err.message : fallback;
}

// useWorkoutFormActions centralizes create/update/close workflow for workouts.
export function useWorkoutFormActions({
  currentUserId,
  workoutDirty,
  setSelectedWorkoutId,
  setEditingWorkout,
  setWorkoutDirty,
  setShowWorkoutForm,
  setWorkouts,
  reloadWorkouts,
  askConfirm,
  notify,
}: UseWorkoutFormActionsArgs) {
  // saveWorkout creates a new workout for the current user.
  const saveWorkout = async (payload: {
    name: string;
    steps: WorkoutStep[];
  }) => {
    // Guard: require an authenticated user before saving.
    if (!currentUserId) {
      await notify?.("You must be logged in to save workouts.");
      return;
    }

    try {
      const created = await createWorkout({
        userId: currentUserId,
        name: payload.name,
        steps: payload.steps,
      });

      // Prepend the new workout to local state.
      setWorkouts((prev) => (prev ? [created, ...prev] : [created]));
      setSelectedWorkoutId(created.id);
      setEditingWorkout(created);
      setWorkoutDirty(false);
    } catch (err) {
      await notify?.(errorMessage(err, "Unable to save workout"));
    }
  };

  // updateWorkout persists updates and refreshes local cache.
  const updateWorkout = async (payload: {
    id: string;
    name: string;
    steps: WorkoutStep[];
  }) => {
    // Guard: require an authenticated user before updating.
    if (!currentUserId) {
      await notify?.("You must be logged in to update workouts.");
      return;
    }

    try {
      const updated = await updateWorkoutApi(payload.id, {
        userId: currentUserId,
        name: payload.name,
        steps: payload.steps,
      });

      // Prefer fetching fresh data if the API response is partial.
      let fresh = updated;
      try {
        fresh = await getWorkout(updated.id);
      } catch {
        // ignore and fallback to update response
      }

      setWorkouts((prev) =>
        prev ? prev.map((w) => (w.id === fresh.id ? fresh : w)) : [fresh],
      );

      reloadWorkouts?.();

      setSelectedWorkoutId(fresh.id);
      setEditingWorkout(fresh);
      setWorkoutDirty(false);
    } catch (err) {
      await notify?.(errorMessage(err, "Unable to update workout"));
    }
  };

  // closeWorkoutModal hides the modal and guards against unsaved edits.
  const closeWorkoutModal = async () => {
    if (workoutDirty) {
      const discard = await askConfirm(
        "You have unsaved changes. Close without saving?",
      );
      if (!discard) return;
    }
    setShowWorkoutForm(false);
    setEditingWorkout(null);
    setWorkoutDirty(false);
  };

  return { saveWorkout, updateWorkout, closeWorkoutModal };
}
