import type {
  CatalogExercise,
  TrainngHistoryItem,
  TrainngState,
  TrainngStepLog,
  SoundOption,
  User,
  Workout,
  WorkoutStep,
  Template,
} from "./types";
import { withBasePath } from "./utils/basePath";

type AppConfig = {
  authHeaderEnabled: boolean;
  allowRegistration: boolean;
  version: string;
  commit: string;
};

let useLocalUserHeader = true;

// setAuthHeaderEnabled toggles local header usage based on proxy auth.
export const setAuthHeaderEnabled = (enabled: boolean) => {
  useLocalUserHeader = !enabled;
};

// request wraps fetch with JSON handling, error surfacing, and user header.
async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const userId = localStorage.getItem("motus:userId") || "";
  const res = await fetch(withBasePath(path), {
    headers: {
      "Content-Type": "application/json",
      ...(useLocalUserHeader && userId ? { "X-User-ID": userId } : {}),
      ...(init?.headers || {}),
    },
    ...init,
  });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = await res.json();
      message = body.error || message;
    } catch {
      // ignore
    }
    throw new Error(message);
  }
  if (res.status === 204) {
    // @ts-expect-error allow void
    return null;
  }
  return res.json() as Promise<T>;
}

// getConfig returns the runtime API configuration.
export async function getConfig(): Promise<AppConfig> {
  return request("/api/config");
}

// getCurrentUser resolves the authenticated user.
export async function getCurrentUser(): Promise<User> {
  return request("/api/me");
}

// listUsers fetches all users for the admin view.
export async function listUsers(): Promise<User[]> {
  return request("/api/users");
}

// createUser creates a new user with the given email.
export async function createUser(
  email: string,
  password?: string,
): Promise<User> {
  return request("/api/users", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

// loginUser authenticates a local user.
export async function loginUser(
  email: string,
  password: string,
): Promise<User> {
  return request("/api/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

// changePassword updates the current user's password.
export async function changePassword(
  currentPassword: string,
  newPassword: string,
): Promise<void> {
  return request("/api/me/password", {
    method: "PUT",
    body: JSON.stringify({ currentPassword, newPassword }),
  });
}

// updateUserName changes the current user's display name.
export async function updateUserName(name: string): Promise<void> {
  return request("/api/me/name", {
    method: "PUT",
    body: JSON.stringify({ name }),
  });
}

// updateUserAdmin toggles admin flag for a user.
export async function updateUserAdmin(
  userId: string,
  isAdmin: boolean,
): Promise<void> {
  return request(`/api/users/${userId}/admin`, {
    method: "PUT",
    body: JSON.stringify({ isAdmin }),
  });
}

// listWorkouts returns all workouts for a user.
export async function listWorkouts(userId: string): Promise<Workout[]> {
  return request(`/api/users/${encodeURIComponent(userId)}/workouts`);
}

// getWorkout fetches a single workout by id.
export async function getWorkout(id: string): Promise<Workout> {
  return request(`/api/workouts/${id}`);
}

// exportWorkout fetches a workout JSON payload for sharing.
export async function exportWorkout(id: string): Promise<Workout> {
  return request(`/api/workouts/${id}/export`);
}

// createWorkout persists a new workout.
export async function createWorkout(payload: {
  userId: string;
  name: string;
  steps: WorkoutStep[];
}): Promise<Workout> {
  return request(`/api/users/${encodeURIComponent(payload.userId)}/workouts`, {
    method: "POST",
    body: JSON.stringify({ name: payload.name, steps: payload.steps }),
  });
}

// updateWorkout updates an existing workout.
export async function updateWorkout(
  workoutId: string,
  payload: { userId: string; name: string; steps: WorkoutStep[] },
): Promise<Workout> {
  return request(`/api/workouts/${workoutId}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

// deleteWorkout removes a workout.
export async function deleteWorkout(workoutId: string): Promise<void> {
  return request(`/api/workouts/${workoutId}`, { method: "DELETE" });
}

// importWorkout creates a workout from an exported JSON payload.
export async function importWorkout(payload: {
  userId?: string;
  workout: Workout;
}): Promise<Workout> {
  return request("/api/workouts/import", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

// shareTemplate makes a workout available as a template.
export async function shareTemplate(workoutId: string, name?: string) {
  return request("/api/templates", {
    method: "POST",
    body: JSON.stringify({ workoutId, name }),
  });
}

// listExercises returns all exercises.
export async function listExercises(): Promise<CatalogExercise[]> {
  return request("/api/exercises");
}

// backfillExercises promotes workout exercises into the core catalog.
export async function backfillExercises(): Promise<void> {
  return request("/api/exercises/backfill", { method: "POST" });
}

// createExercise adds a new exercise.
export async function createExercise(
  name: string,
  isCore = false,
): Promise<CatalogExercise> {
  return request("/api/exercises", {
    method: "POST",
    body: JSON.stringify({ name, isCore }),
  });
}

// updateExercise renames an exercise.
export async function updateExercise(
  id: string,
  name: string,
): Promise<CatalogExercise> {
  return request(`/api/exercises/${id}`, {
    method: "PUT",
    body: JSON.stringify({ name }),
  });
}

// deleteExercise removes an exercise.
export async function deleteExercise(id: string) {
  return request(`/api/exercises/${id}`, { method: "DELETE" });
}

// listSounds returns available sound options.
export async function listSounds(): Promise<SoundOption[]> {
  return request("/api/sounds");
}

// startTraining creates a new training for a workout.
export async function startTrain(workoutId: string): Promise<TrainngState> {
  const res = await request<{ trainingId: string; state: TrainngState }>(
    "/api/trainings",
    {
      method: "POST",
      body: JSON.stringify({ workoutId }),
    },
  );
  return res.state;
}

// logTrainingCompletion records a completed training.
export async function logTrainingCompletion(payload: {
  trainingId: string;
  workoutId: string;
  workoutName?: string;
  userId: string;
  startedAt: string;
  completedAt: string;
  steps?: Array<{
    id?: string;
    name: string;
    type: string;
    estimatedSeconds?: number;
    elapsedMillis?: number;
  }>;
}) {
  return request("/api/trainings/complete", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

// listTrainingHistory returns all completed trainings for a user.
export async function listTrainingHistory(
  userId: string,
): Promise<TrainngHistoryItem[]> {
  return request(`/api/users/${encodeURIComponent(userId)}/trainings/history`);
}

// getTrainingSteps fetches stored per-step timings for a training.
export async function getTrainingSteps(
  trainingId: string,
): Promise<TrainngStepLog[]> {
  return request(`/api/trainings/${encodeURIComponent(trainingId)}/steps`);
}

// listTemplates returns all templates.
export async function listTemplates(): Promise<Template[]> {
  return request("/api/templates");
}

// applyTemplate clones a template into a workout.
export async function applyTemplate(
  templateId: string,
  payload: { userId: string; name?: string },
): Promise<Workout> {
  return request(`/api/templates/${templateId}/apply`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}
