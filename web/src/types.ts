// Exercise describes a single exercise entry inside a workout step.
export type Exercise = {
  exerciseId?: string;
  name: string;
  amount?: string;
  weight?: string;
  // For timed sets, durationSeconds and transitionSeconds are derived from amount/weight inputs.
  durationSeconds?: number;
  transitionSeconds?: number;
};

// PauseOptions configures pause step behavior.
export type PauseOptions = {
  autoAdvance?: boolean;
};

// CatalogExercise represents an exercise in the shared catalog.
export type CatalogExercise = {
  id: string;
  name: string;
  isCore?: boolean;
  ownerUserId?: string;
  createdAt?: string;
};

// Template describes a reusable workout template.
export type Template = {
  id: string;
  name: string;
  steps: WorkoutStep[];
};

// WorkoutStep defines a single step inside a workout.
export type WorkoutStep = {
  id?: string;
  order?: number;
  type: "set" | "pause" | "timed";
  name: string;
  weight?: string;
  estimatedSeconds?: number;
  soundKey?: string;
  soundUrl?: string;
  duration?: string;
  exercises?: Exercise[];
  pauseOptions?: PauseOptions;
  repeatCount?: number;
  repeatRestSeconds?: number;
  repeatRestAfterLast?: boolean;
  repeatRestSoundKey?: string;
  repeatRestAutoAdvance?: boolean;
};

// View represents the active app section.
export type View =
  | "sessions"
  | "login"
  | "workouts"
  | "profile"
  | "history"
  | "exercises"
  | "templates"
  | "admin";

// Workout represents a full workout with steps.
export type Workout = {
  id: string;
  userId: string;
  name: string;
  createdAt?: string;
  steps: WorkoutStep[];
};

// User describes a Motus account.
export type User = {
  id: string;
  name: string;
  isAdmin?: boolean;
  createdAt: string;
};

// SessionStepState captures a live session step.
export type SessionStepState = {
  id?: string;
  name: string;
  type: "set" | "pause" | "timed";
  estimatedSeconds?: number;
  soundUrl?: string;
  soundKey?: string;
  exercises?: Exercise[];
  elapsedMillis: number;
  completed: boolean;
  current: boolean;
  running: boolean;
  soundPlayed?: boolean;
  pauseOptions?: PauseOptions;
  loopIndex?: number;
  loopTotal?: number;
};

// SessionStepLog stores a completed step timing.
export type SessionStepLog = {
  id: string;
  sessionId: string;
  stepOrder: number;
  type: string;
  name: string;
  estimatedSeconds: number;
  elapsedMillis: number;
};

// SessionState tracks the active workout session.
export type SessionState = {
  sessionId: string;
  workoutId: string;
  workoutName?: string;
  userId: string;
  currentIndex: number;
  running: boolean;
  runningSince?: number | null;
  done: boolean;
  startedAt?: string | null;
  completedAt?: string | null;
  logged?: boolean;
  steps: SessionStepState[];
};

// SessionHistoryItem summarizes a completed session with optional steps.
export type SessionHistoryItem = {
  id: string;
  sessionId: string;
  workoutId: string;
  workoutName?: string;
  userId: string;
  startedAt?: string;
  completedAt?: string;
  steps?: SessionStepLog[];
};

// SoundOption describes an available sound effect.
export type SoundOption = {
  key: string;
  label: string;
  file: string;
  leadSeconds?: number;
};
