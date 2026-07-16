import type {
  CatalogExercise,
  TrainingHistoryItem,
  TrainingState,
  TrainingStepLog,
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

// setAuthHeaderEnabled toggles local header usage based on proxy auth.
export const setAuthHeaderEnabled = (_enabled: boolean) => {};

// request wraps fetch with JSON handling, error surfacing, and user header.
async function requestResponse(path: string, init?: RequestInit): Promise<Response> {
  const res = await fetch(withBasePath(path), {
	credentials: "same-origin",
    headers: {
      "Content-Type": "application/json",
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
  return res;
}

async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await requestResponse(path, init);
  return response.json() as Promise<T>;
}

async function requestVoid(path: string, init?: RequestInit): Promise<void> {
  await requestResponse(path, init);
}

// getConfig returns the runtime API configuration.
export async function getConfig(): Promise<AppConfig> {
  return requestJSON("/api/config");
}

// getCurrentUser resolves the authenticated user.
export async function getCurrentUser(): Promise<User> {
  return requestJSON("/api/me");
}

// listUsers fetches all users for the admin view.
export async function listUsers(): Promise<User[]> {
  return requestJSON("/api/users");
}

// createUser creates a new user with the given email.
export async function createUser(
  email: string,
  password?: string,
): Promise<User> {
  return requestJSON("/api/users", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

// loginUser authenticates a local user.
export async function loginUser(
  email: string,
  password: string,
): Promise<User> {
  return requestJSON("/api/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

// changePassword updates the current user's password.
export async function changePassword(
  currentPassword: string,
  newPassword: string,
): Promise<void> {
  return requestVoid("/api/me/password", {
    method: "PUT",
    body: JSON.stringify({ currentPassword, newPassword }),
  });
}

// updateUserName changes the current user's display name.
export async function updateUserName(name: string): Promise<void> {
  return requestVoid("/api/me/name", {
    method: "PUT",
    body: JSON.stringify({ name }),
  });
}

// updateUserAdmin toggles admin flag for a user.
export async function updateUserAdmin(
  userId: string,
  isAdmin: boolean,
): Promise<void> {
  return requestVoid(`/api/users/${userId}/admin`, {
    method: "PUT",
    body: JSON.stringify({ isAdmin }),
  });
}

// listWorkouts returns all workouts for a user.
export async function listWorkouts(userId: string): Promise<Workout[]> {
  return requestJSON(`/api/users/${encodeURIComponent(userId)}/workouts`);
}

// getWorkout fetches a single workout by id.
export async function getWorkout(id: string): Promise<Workout> {
  return requestJSON(`/api/workouts/${id}`);
}

// exportWorkout fetches a workout JSON payload for sharing.
export async function exportWorkout(id: string): Promise<Workout> {
  return requestJSON(`/api/workouts/${id}/export`);
}

// createWorkout persists a new workout.
export async function createWorkout(payload: {
  userId: string;
  name: string;
  steps: WorkoutStep[];
}): Promise<Workout> {
  return requestJSON(`/api/users/${encodeURIComponent(payload.userId)}/workouts`, {
    method: "POST",
    body: JSON.stringify({ name: payload.name, steps: payload.steps }),
  });
}

// updateWorkout updates an existing workout.
export async function updateWorkout(
  workoutId: string,
  payload: { userId: string; name: string; steps: WorkoutStep[] },
): Promise<Workout> {
  return requestJSON(`/api/workouts/${workoutId}`, {
    method: "PUT",
    body: JSON.stringify(payload),
  });
}

// deleteWorkout removes a workout.
export async function deleteWorkout(workoutId: string): Promise<void> {
  return requestVoid(`/api/workouts/${workoutId}`, { method: "DELETE" });
}

// importWorkout creates a workout from an exported JSON payload.
export async function importWorkout(payload: {
  userId?: string;
  workout: Workout;
}): Promise<Workout> {
  return requestJSON("/api/workouts/import", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

// shareTemplate makes a workout available as a template.
export async function shareTemplate(workoutId: string, name?: string) {
  return requestJSON("/api/templates", {
    method: "POST",
    body: JSON.stringify({ workoutId, name }),
  });
}

// listExercises returns all exercises.
export async function listExercises(): Promise<CatalogExercise[]> {
  return requestJSON("/api/exercises");
}

// backfillExercises promotes workout exercises into the core catalog.
export async function backfillExercises(): Promise<void> {
  return requestVoid("/api/exercises/backfill", { method: "POST" });
}

// createExercise adds a new exercise.
export async function createExercise(
  name: string,
  isCore = false,
): Promise<CatalogExercise> {
  return requestJSON("/api/exercises", {
    method: "POST",
    body: JSON.stringify({ name, isCore }),
  });
}

// updateExercise renames an exercise.
export async function updateExercise(
  id: string,
  name: string,
  hasSides?: boolean,
): Promise<CatalogExercise> {
  return requestJSON(`/api/exercises/${id}`, {
    method: "PUT",
    body: JSON.stringify({ name, hasSides }),
  });
}

// deleteExercise removes an exercise.
export async function deleteExercise(id: string) {
  return requestVoid(`/api/exercises/${id}`, { method: "DELETE" });
}

// listSounds returns available sound options.
export async function listSounds(): Promise<SoundOption[]> {
  return requestJSON("/api/sounds");
}

// startTraining creates a new training for a workout.
export async function startTraining(workoutId: string): Promise<TrainingState> {
  const res = await requestJSON<{ trainingId: string; state: TrainingState }>(
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
  return requestJSON("/api/trainings/complete", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

// listTrainingHistory returns all completed trainings for a user.
export async function listTrainingHistory(
  userId: string,
): Promise<TrainingHistoryItem[]> {
  return requestJSON(`/api/users/${encodeURIComponent(userId)}/trainings/history`);
}

// getTrainingSteps fetches stored per-step timings for a training.
export async function getTrainingSteps(
  trainingId: string,
): Promise<TrainingStepLog[]> {
  return requestJSON(`/api/trainings/${encodeURIComponent(trainingId)}/steps`);
}

// listTemplates returns all templates.
export async function listTemplates(): Promise<Template[]> {
  return requestJSON("/api/templates");
}

// applyTemplate clones a template into a workout.
export async function applyTemplate(
  templateId: string,
  payload: { userId: string; name?: string },
): Promise<Workout> {
  return requestJSON(`/api/templates/${templateId}/apply`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}
