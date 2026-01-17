import { useCallback } from "react";
import { MESSAGES, toErrorMessage } from "../utils/messages";

import { changePassword, exportWorkout, importWorkout } from "../api";
import type { Workout } from "../types";

// UseProfileActionsArgs wires profile and transfer actions.
type UseProfileActionsArgs = {
  currentUserId: string | null;
  exportWorkoutId: string;
  setSelectedWorkoutId: (id: string) => void;
  setWorkouts: (updater: (prev: Workout[] | null) => Workout[] | null) => void;
  showToast: (message: string) => void;
  notify: (message: string) => Promise<void>;
};

// useProfileActions provides profile settings handlers.
export function useProfileActions({
  currentUserId,
  exportWorkoutId,
  setSelectedWorkoutId,
  setWorkouts,
  showToast,
  notify,
}: UseProfileActionsArgs) {
  // exportSelectedWorkout downloads the selected workout JSON.
  const exportSelectedWorkout = useCallback(async () => {
    if (!exportWorkoutId) {
      // Guard: require a selection before exporting.
      await notify("Select a workout to export.");
      return;
    }
    try {
      const workout = await exportWorkout(exportWorkoutId);
      const blob = new Blob([JSON.stringify(workout, null, 2)], {
        type: "application/json",
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `${workout.name || "workout"}.json`;
      a.click();
      URL.revokeObjectURL(url);
      showToast("Workout exported.");
    } catch (err) {
      await notify(toErrorMessage(err, MESSAGES.exportWorkoutFailed));
    }
  }, [exportWorkoutId, notify, showToast]);

  // importWorkoutFile uploads a workout JSON payload.
  const importWorkoutFile = useCallback(
    async (file: File) => {
      try {
        const raw = await file.text();
        const parsed = JSON.parse(raw);
        // Accept either { workout: {...} } or a raw workout export.
        const workoutPayload = parsed.workout ? parsed.workout : parsed;
        if (!workoutPayload?.name || !workoutPayload?.steps) {
          await notify("Invalid workout JSON.");
          return;
        }
        const created = await importWorkout({
          userId: currentUserId || undefined,
          workout: workoutPayload,
        });
        setWorkouts((prev) => (prev ? [created, ...prev] : [created]));
        setSelectedWorkoutId(created.id);
        showToast("Workout imported.");
      } catch (err) {
        await notify(toErrorMessage(err, MESSAGES.importWorkoutFailed));
      }
    },
    [currentUserId, notify, setSelectedWorkoutId, setWorkouts, showToast],
  );

  // updatePassword changes the current user's password.
  const updatePassword = useCallback(
    async (currentPassword: string, newPassword: string) => {
      try {
        await changePassword(currentPassword, newPassword);
        await notify("Password updated.");
      } catch (err) {
        await notify(toErrorMessage(err, MESSAGES.updatePasswordFailed));
      }
    },
    [notify],
  );

  return { exportSelectedWorkout, importWorkoutFile, updatePassword };
}
