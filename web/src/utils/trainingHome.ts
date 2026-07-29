import type { TrainingHistoryItem, Workout } from "../types";

export function selectQuickWorkouts(
  workouts: Workout[],
  history: TrainingHistoryItem[],
  favoriteIds: string[],
  limit = 4,
): Workout[] {
  const recentIds = [...history]
    .sort((a, b) =>
      String(b.completedAt || "").localeCompare(String(a.completedAt || "")),
    )
    .map((item) => item.workoutId);
  const priority = [...favoriteIds, ...recentIds].filter(
    (id, index, ids) => ids.indexOf(id) === index,
  );
  return priority
    .map((id) => workouts.find((workout) => workout.id === id))
    .filter((workout): workout is Workout => Boolean(workout))
    .slice(0, limit);
}
