import { useCallback } from "react";
import { MESSAGES, PROMPTS, toErrorMessage } from "../utils/messages";
import { UI_TEXT } from "../utils/uiText";

import { changePassword, exportWorkout, importWorkout } from "../api";
import type { Workout } from "../types";
import { withBasePath } from "../utils/basePath";

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
  const importWorkoutPayload = useCallback(
    async (workoutPayload: Workout) => {
      if (!workoutPayload?.name || !workoutPayload?.steps) {
        await notify(UI_TEXT.toasts.invalidWorkoutJson);
        return;
      }
      const created = await importWorkout({
        userId: currentUserId || undefined,
        workout: workoutPayload,
      });
      setWorkouts((prev) => (prev ? [created, ...prev] : [created]));
      setSelectedWorkoutId(created.id);
      showToast(UI_TEXT.toasts.workoutImported);
    },
    [currentUserId, notify, setSelectedWorkoutId, setWorkouts, showToast],
  );

  // exportSelectedWorkout downloads the selected workout JSON.
  const exportSelectedWorkout = useCallback(async () => {
    if (!exportWorkoutId) {
      // Guard: require a selection before exporting.
      await notify(PROMPTS.selectWorkoutToExport);
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
      showToast(UI_TEXT.toasts.workoutExported);
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
        await importWorkoutPayload(workoutPayload);
      } catch (err) {
        await notify(toErrorMessage(err, MESSAGES.importWorkoutFailed));
      }
    },
    [importWorkoutPayload, notify],
  );

  const importTestWorkout = useCallback(async () => {
    try {
      const response = await fetch(withBasePath("/test-workout.json"));
      if (!response.ok) throw new Error("Unable to load test workout");
      await importWorkoutPayload((await response.json()) as Workout);
    } catch (err) {
      await notify(toErrorMessage(err, MESSAGES.importWorkoutFailed));
    }
  }, [importWorkoutPayload, notify]);

  // updatePassword changes the current user's password.
  const updatePassword = useCallback(
    async (currentPassword: string, newPassword: string) => {
      try {
        await changePassword(currentPassword, newPassword);
        await notify(UI_TEXT.toasts.passwordUpdated);
      } catch (err) {
        await notify(toErrorMessage(err, MESSAGES.updatePasswordFailed));
      }
    },
    [notify],
  );

  return {
    exportSelectedWorkout,
    importWorkoutFile,
    importTestWorkout,
    updatePassword,
  };
}
