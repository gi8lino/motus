import type { StepType } from "./utils/step";

// ThemeMode controls the app theme selection.
export type ThemeMode = "auto" | "dark" | "light";

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

// TrainStepState captures a live training step.
export type TrainingStepState = WorkoutStep & {
  elapsedMillis: number;
  completed: boolean;
  current: boolean;
  running: boolean;

  soundPlayed?: boolean;

  subsetId?: string;
  subsetLabel?: string;
  hasMultipleSubsets?: boolean;
  superset?: boolean;
  setName?: string;
  subsetEstimatedSeconds?: number;
};

// TrainStepLog stores a completed step timing.
export type TrainingStepLog = {
  id: string;
  trainingId: string;
  stepOrder: number;
  type: string;
  name: string;
  estimatedSeconds: number;
  elapsedMillis: number;
};

// TrainingState tracks the active workout training.
export type TrainingState = {
  trainingId: string;
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

  steps: TrainingStepState[];
};

// TrainingHistoryItem summarizes a completed training with optional steps.
export type TrainingHistoryItem = {
  id: string;
  trainingId: string;
  workoutId: string;
  workoutName?: string;
  userId: string;
  startedAt?: string;
  completedAt?: string;
  steps?: TrainingStepLog[];
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
