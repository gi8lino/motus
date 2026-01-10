export const EXERCISE_TYPE_REP = "rep";
export const EXERCISE_TYPE_STOPWATCH = "stopwatch";
export const EXERCISE_TYPE_COUNTDOWN = "countdown";
export const EXERCISE_TYPE_TIMED = "timed"; // legacy alias

export type ExerciseKind =
  | typeof EXERCISE_TYPE_REP
  | typeof EXERCISE_TYPE_STOPWATCH
  | typeof EXERCISE_TYPE_COUNTDOWN
  | typeof EXERCISE_TYPE_TIMED;

export function normalizeExerciseType(value?: string): ExerciseKind {
  const token = (value || "").trim().toLowerCase();
  if (token === EXERCISE_TYPE_COUNTDOWN) return EXERCISE_TYPE_COUNTDOWN;
  if (token === EXERCISE_TYPE_STOPWATCH || token === EXERCISE_TYPE_TIMED)
    return EXERCISE_TYPE_STOPWATCH;
  return EXERCISE_TYPE_REP;
}

export function isDurationExercise(value?: string): boolean {
  const type = normalizeExerciseType(value);
  return type === EXERCISE_TYPE_STOPWATCH || type === EXERCISE_TYPE_COUNTDOWN;
}
