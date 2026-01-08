import { useCallback } from "react";

import { deleteWorkout, shareTemplate } from "../api";
import type { Workout } from "../types";

// UseWorkoutActionsArgs wires workout list actions.
type UseWorkoutActionsArgs = {
  workouts: Workout[];
  editingWorkout: Workout | null;
  selectedWorkoutId: string | null;
  setEditingWorkout: (workout: Workout | null) => void;
  setShowWorkoutForm: (open: boolean) => void;
  setSelectedWorkoutId: (id: string | null) => void;
  setWorkouts: (updater: (prev: Workout[] | null) => Workout[] | null) => void;
  askConfirm: (message: string) => Promise<boolean>;
  askPrompt: (message: string, defaultValue?: string) => Promise<string | null>;
  notify: (message: string) => Promise<void>;
  templatesReload: () => void;
};

// useWorkoutActions provides handlers for workout list actions.
export function useWorkoutActions({
  workouts,
  editingWorkout,
  selectedWorkoutId,
  setEditingWorkout,
  setShowWorkoutForm,
  setSelectedWorkoutId,
  setWorkouts,
  askConfirm,
  askPrompt,
  notify,
  templatesReload,
}: UseWorkoutActionsArgs) {
  // newWorkout opens the workout creation modal.
  const newWorkout = useCallback(() => {
    setEditingWorkout(null);
    setShowWorkoutForm(true);
  }, [setEditingWorkout, setShowWorkoutForm]);

  // editWorkout opens the modal for the selected workout.
  const editWorkout = useCallback(
    (workoutId: string) => {
      const found = workouts.find((w) => w.id === workoutId);
      if (found) {
        setEditingWorkout(found);
        setShowWorkoutForm(true);
      }
    },
    [workouts, setEditingWorkout, setShowWorkoutForm],
  );

  // editWorkoutFromList selects a workout and opens the modal.
  const editWorkoutFromList = useCallback(
    (workoutId: string) => {
      setSelectedWorkoutId(workoutId);
      editWorkout(workoutId);
      setShowWorkoutForm(true);
    },
    [editWorkout, setSelectedWorkoutId, setShowWorkoutForm],
  );

  // removeWorkout deletes a workout after confirmation.
  const removeWorkout = useCallback(
    async (workoutId: string) => {
      const target = workouts.find((w) => w.id === workoutId);
      const name = target?.name || "this workout";
      const ok = await askConfirm(`Delete ${name}? This cannot be undone.`);
      if (!ok) return;
      try {
        // Delete on the server before updating local state.
        await deleteWorkout(workoutId);
        setWorkouts((prev) =>
          prev ? prev.filter((w) => w.id !== workoutId) : prev,
        );
        if (selectedWorkoutId === workoutId) {
          setSelectedWorkoutId(null);
        }
        if (editingWorkout?.id === workoutId) {
          setEditingWorkout(null);
          setShowWorkoutForm(false);
        }
      } catch (err: any) {
        await notify(err.message || "Unable to delete workout");
      }
    },
    [
      workouts,
      askConfirm,
      setWorkouts,
      editingWorkout?.id,
      selectedWorkoutId,
      setEditingWorkout,
      setShowWorkoutForm,
      setSelectedWorkoutId,
      notify,
    ],
  );

  // shareWorkout publishes a workout as a template.
  const shareWorkout = useCallback(
    async (workoutId: string) => {
      const name = await askPrompt("Template name (optional)");
      if (name === null) return;
      try {
        // Share then refresh template list for immediate visibility.
        await shareTemplate(workoutId, name.trim());
        templatesReload();
        await notify("Template shared.");
      } catch (err: any) {
        await notify(err.message || "Unable to share template");
      }
    },
    [askPrompt, notify, templatesReload],
  );

  return {
    newWorkout,
    editWorkout,
    editWorkoutFromList,
    removeWorkout,
    shareWorkout,
  };
}
