import type { WorkoutStep } from "../types";

// workoutWriteSteps returns the strict wire representation accepted by the API.
// Editor-only identifiers and derived display fields must not cross this boundary.
export function workoutWriteSteps(steps: WorkoutStep[]) {
  return steps.map((step) => ({
    type: step.type,
    name: step.name,
    duration: step.duration ?? "",
    estimatedSeconds: step.estimatedSeconds ?? 0,
    soundKey: step.soundKey ?? "",
    subsets: (step.subsets ?? []).map((subset) => ({
      name: subset.name,
      duration: subset.duration ?? "",
      soundKey: subset.soundKey ?? "",
      superset: Boolean(subset.superset),
      exercises: (subset.exercises ?? []).map((exercise) => ({
        exerciseId: exercise.exerciseId ?? "",
        name: exercise.name,
        type: exercise.type ?? "rep",
        reps: exercise.reps ?? "",
        weight: exercise.weight ?? "",
        duration: exercise.duration ?? "",
        soundKey: exercise.soundKey ?? "",
        side: exercise.side ?? "not_applicable",
      })),
    })),
    pauseOptions: step.pauseOptions ?? {},
    repeatCount: step.repeatCount ?? 1,
    repeatRestSeconds: step.repeatRestSeconds ?? 0,
    repeatRestAfterLast: Boolean(step.repeatRestAfterLast),
    repeatRestSoundKey: step.repeatRestSoundKey ?? "",
    repeatRestAutoAdvance: Boolean(step.repeatRestAutoAdvance),
    repeatRestName: step.repeatRestName ?? "",
  }));
}
