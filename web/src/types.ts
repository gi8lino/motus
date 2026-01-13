// Exercise describes a single exercise entry inside a workout step.
export type Exercise = {
  exerciseId?: string;
  name: string;
  type?: "rep" | "stopwatch" | "countdown";
  reps?: string;
  weight?: string;
  duration?: string;
  soundKey?: string;
};

// PauseOptions configures pause step behavior.
export type PauseOptions = {
  autoAdvance?: boolean;
};

// WorkoutSubset represents a logical subset inside a set step.
export type WorkoutSubset = {
  id?: string;
  name: string;
  duration?: string;
  estimatedSeconds?: number;
  soundKey?: string;
  superset?: boolean;
  exercises?: Exercise[];
};

// WorkoutStep defines a single step inside a workout.
import type { StepType } from "./utils/step";

export type WorkoutStep = {
  id?: string;
  order?: number;
  type: StepType;
  name: string;
  duration?: string;
  estimatedSeconds?: number;
  soundKey?: string;
  soundUrl?: string;
  exercises?: Exercise[];
  subsets?: WorkoutSubset[];
  pauseOptions?: PauseOptions;
  autoAdvance?: boolean;
  repeatCount?: number;
  repeatRestSeconds?: number;
  repeatRestAfterLast?: boolean;
  repeatRestSoundKey?: string;
  repeatRestAutoAdvance?: boolean;
  loopIndex?: number;
  loopTotal?: number;
};

// Workout represents a full workout with steps.
export type Workout = {
  id: string;
  userId: string;
  name: string;
  createdAt?: string;
  isTemplate?: boolean;
  steps: WorkoutStep[];
};

// Template describes a reusable workout template.
export type Template = Workout;

// View represents the active app section.
export type View =
  | "train"
  | "login"
  | "workouts"
  | "profile"
  | "history"
  | "exercises"
  | "templates"
  | "admin";

// User describes a Motus account.
export type User = {
  id: string;
  name: string;
  isAdmin?: boolean;
  createdAt: string;
};

// SessionStepState captures a live session step.
export type SessionStepState = WorkoutStep & {
  id?: string;
  elapsedMillis: number;
  completed: boolean;
  current: boolean;
  running: boolean;
  soundUrl?: string;
  soundKey?: string;
  soundPlayed?: boolean;
  pauseOptions?: PauseOptions;
  subsetId?: string;
  subsetLabel?: string;
  hasMultipleSubsets?: boolean;
  superset?: boolean;
  setName?: string;
  subsetEstimatedSeconds?: number;
  autoAdvance?: boolean;
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

// CatalogExercise describes an exercise stored in the shared catalog.
export type CatalogExercise = {
  id: string;
  name: string;
  ownerUserId?: string;
  isCore?: boolean;
  createdAt?: string;
};

// AskConfirmOptions configures confirm dialog labels.
export type AskConfirmOptions = {
  title?: string;
  confirmLabel?: string;
  cancelLabel?: string;
};
