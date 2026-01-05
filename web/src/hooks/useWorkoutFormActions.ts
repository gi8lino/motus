import {
  createWorkout,
  getWorkout,
  updateWorkout as updateWorkoutApi,
} from "../api";
import type { Workout, WorkoutStep } from "../types";

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
  askConfirm: (message: string) => Promise<boolean>;
};

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
}: UseWorkoutFormActionsArgs) {
  // saveWorkout creates a new workout for the current user.
  const saveWorkout = async (payload: {
    name: string;
    steps: WorkoutStep[];
  }) => {
    // Guard: require an authenticated user before saving.
    if (!currentUserId) return;
    const created = await createWorkout({
      userId: currentUserId,
      name: payload.name,
      steps: payload.steps,
    });
    // Prepend the new workout to local state.
    setWorkouts((prev) => (prev ? [created, ...prev] : [created]));
    setSelectedWorkoutId(created.id);
    setEditingWorkout(null);
    setWorkoutDirty(false);
  };

  // updateWorkout persists updates and refreshes local cache.
  const updateWorkout = async (payload: {
    id: string;
    name: string;
    steps: WorkoutStep[];
  }) => {
    // Guard: require an authenticated user before updating.
    if (!currentUserId) return;
    const updated = await updateWorkoutApi(payload.id, {
      userId: currentUserId,
      name: payload.name,
      steps: payload.steps,
    });
    let fresh = updated;
    try {
      fresh = await getWorkout(updated.id);
    } catch {
      // ignore and fallback to response
    }
    setWorkouts((prev) =>
      prev ? prev.map((w) => (w.id === fresh.id ? fresh : w)) : [fresh],
    );
    reloadWorkouts?.();
    setSelectedWorkoutId(fresh.id);
    setEditingWorkout(fresh);
    setWorkoutDirty(false);
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
