import { useMemo } from "react";
import {
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
};

// useWorkoutsData loads shared datasets and keeps derived views in sync.
export function useWorkoutsData({ currentUserId }: UseWorkoutsDataArgs) {
  const users = useDataLoader<User[]>(listUsers, []);
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
  const currentUser = useMemo(
    () => users.data?.find((u) => u.id === currentUserId) || null,
    [users.data, currentUserId],
  );

  return {
    users,
    sounds,
    workouts,
    history,
    templates,
    activeWorkouts,
    currentUser,
  };
}
