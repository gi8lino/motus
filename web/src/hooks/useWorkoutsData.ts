import { useMemo } from "react";
import {
  getCurrentUser,
  listTrainingHistory,
  listSounds,
  listTemplates,
  listUsers,
  listWorkouts,
} from "../api";
import type {
  TrainingHistoryItem,
  SoundOption,
  Template,
  User,
  Workout,
} from "../types";
import { useDataLoader } from "./useDataLoader";

type UseWorkoutsDataArgs = {
  currentUserId: string | null;
  authHeaderEnabled: boolean;
};

// useWorkoutsData loads shared datasets and keeps derived views in sync.
export function useWorkoutsData({
  currentUserId,
  authHeaderEnabled,
}: UseWorkoutsDataArgs) {
  const canLoadUser = Boolean(authHeaderEnabled || currentUserId);
  const currentUserLoader = useDataLoader<User | null>(
    () => (canLoadUser ? getCurrentUser() : Promise.resolve(null)),
    [canLoadUser],
  );
  const isAdmin = Boolean(currentUserLoader.data?.isAdmin);
  const users = useDataLoader<User[]>(
    () => (isAdmin ? listUsers() : Promise.resolve([])),
    [isAdmin],
  );
  const sounds = useDataLoader<SoundOption[]>(listSounds, []);
  const workouts = useDataLoader<Workout[]>(
    () => (currentUserId ? listWorkouts(currentUserId) : Promise.resolve([])),
    [currentUserId],
  );
  const history = useDataLoader<TrainingHistoryItem[]>(
    () =>
      currentUserId
        ? listTrainingHistory(currentUserId)
        : Promise.resolve([] as TrainingHistoryItem[]),
    [currentUserId],
  );
  const templates = useDataLoader<Template[]>(listTemplates, []);

  const activeWorkouts = workouts.data || [];
  const currentUser = useMemo(() => currentUserLoader.data || null, [
    currentUserLoader.data,
  ]);

  return {
    currentUserLoader,
    users,
    sounds,
    workouts,
    history,
    templates,
    activeWorkouts,
    currentUser,
  };
}
