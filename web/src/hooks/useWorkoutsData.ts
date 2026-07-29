import { useMemo } from "react";
import {
  getCurrentUser,
  listExercises,
  listTrainingHistory,
  listSounds,
  listTemplates,
  listUsers,
  listWorkouts,
} from "../api";
import type {
  CatalogExercise,
  TrainingHistoryItem,
  SoundOption,
  Template,
  User,
  View,
  Workout,
} from "../types";
import { useDataLoader } from "./useDataLoader";

type UseWorkoutsDataArgs = {
  currentUserId: string | null;
  authHeaderEnabled: boolean;
  view: View;
  prefetchedViews: ReadonlySet<View>;
};

// useWorkoutsData loads shared datasets and keeps derived views in sync.
export function useWorkoutsData({
  currentUserId,
  authHeaderEnabled,
  view,
  prefetchedViews,
}: UseWorkoutsDataArgs) {
  const canLoadUser = Boolean(authHeaderEnabled || currentUserId);
  const wants = (candidate: View) =>
    view === candidate || prefetchedViews.has(candidate);
  const needsWorkouts =
    canLoadUser && (wants("train") || wants("workouts") || wants("profile"));
  const needsSounds =
    canLoadUser && (wants("train") || wants("workouts") || wants("profile"));
  const needsExercises =
    canLoadUser && (wants("workouts") || wants("exercises"));
  const needsHistory = canLoadUser && wants("history");
  const needsTemplates = canLoadUser && wants("templates");

  const currentUserLoader = useDataLoader<User | null>(
    () => (canLoadUser ? getCurrentUser() : Promise.resolve(null)),
    [canLoadUser],
    { enabled: canLoadUser, cacheKey: currentUserId },
  );
  const isAdmin = Boolean(currentUserLoader.data?.isAdmin);
  const users = useDataLoader<User[]>(
    () => (isAdmin ? listUsers() : Promise.resolve([])),
    [isAdmin],
    {
      enabled: isAdmin && wants("admin"),
      cacheKey: currentUserId,
    },
  );
  const sounds = useDataLoader<SoundOption[]>(listSounds, [], {
    enabled: needsSounds,
    cacheKey: currentUserId,
  });
  const workouts = useDataLoader<Workout[]>(
    () => listWorkouts(currentUserId!),
    [currentUserId],
    { enabled: needsWorkouts, cacheKey: currentUserId },
  );
  const exercises = useDataLoader<CatalogExercise[]>(listExercises, [], {
    enabled: needsExercises,
    cacheKey: currentUserId,
  });
  const history = useDataLoader<TrainingHistoryItem[]>(
    () => listTrainingHistory(currentUserId!),
    [currentUserId],
    { enabled: needsHistory, cacheKey: currentUserId },
  );
  const templates = useDataLoader<Template[]>(listTemplates, [], {
    enabled: needsTemplates,
    cacheKey: currentUserId,
  });

  const activeWorkouts = workouts.data || [];
  const currentUser = useMemo(
    () => currentUserLoader.data || null,
    [currentUserLoader.data],
  );

  return {
    currentUserLoader,
    users,
    sounds,
    workouts,
    exercises,
    history,
    templates,
    activeWorkouts,
    currentUser,
  };
}
