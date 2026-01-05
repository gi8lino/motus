import { useCallback } from "react";

import { applyTemplate } from "../api";
import type { View, Workout } from "../types";

// UseTemplateActionsArgs wires template application actions.
type UseTemplateActionsArgs = {
  currentUserId: string | null;
  setWorkouts: (updater: (prev: Workout[] | null) => Workout[] | null) => void;
  setSelectedWorkoutId: (id: string | null) => void;
  setShowWorkoutForm: (open: boolean) => void;
  setView: (view: View) => void;
  askPrompt: (message: string, defaultValue?: string) => Promise<string | null>;
  notify: (message: string) => Promise<void>;
};

// useTemplateActions provides handlers for applying templates.
export function useTemplateActions({
  currentUserId,
  setWorkouts,
  setSelectedWorkoutId,
  setShowWorkoutForm,
  setView,
  askPrompt,
  notify,
}: UseTemplateActionsArgs) {
  // applyTemplateToUser clones a shared template into a workout.
  const applyTemplateToUser = useCallback(
    async (templateId: string) => {
      if (!currentUserId) {
        await notify("Select a user first.");
        return;
      }
      const name = await askPrompt("Workout name (optional)");
      if (name === null) return;
      try {
        const workout = await applyTemplate(templateId, {
          userId: currentUserId,
          name: name.trim() || undefined,
        });
        setWorkouts((prev) => (prev ? [workout, ...prev] : [workout]));
        setSelectedWorkoutId(workout.id);
        setView("workouts");
        setShowWorkoutForm(true);
      } catch (err: any) {
        await notify(err.message || "Unable to apply template");
      }
    },
    [
      currentUserId,
      askPrompt,
      notify,
      setWorkouts,
      setSelectedWorkoutId,
      setView,
      setShowWorkoutForm,
    ],
  );

  return { applyTemplateToUser };
}
