export const EXERCISE_TYPE_REP = "rep";
export const EXERCISE_TYPE_STOPWATCH = "stopwatch";
export const EXERCISE_TYPE_COUNTDOWN = "countdown";
export type ExerciseKind =
  | typeof EXERCISE_TYPE_REP
  | typeof EXERCISE_TYPE_STOPWATCH
  | typeof EXERCISE_TYPE_COUNTDOWN;

// normalizeExerciseType coerces raw values into a known exercise kind.
export function normalizeExerciseType(value?: string): ExerciseKind {
  const token = (value || "").trim().toLowerCase();
  if (token === EXERCISE_TYPE_COUNTDOWN) return EXERCISE_TYPE_COUNTDOWN;
  if (token === EXERCISE_TYPE_STOPWATCH) return EXERCISE_TYPE_STOPWATCH;
  return EXERCISE_TYPE_REP;
}

// isDurationExercise reports whether an exercise uses a stopwatch/countdown timer.
export function isDurationExercise(value?: string): boolean {
  const type = normalizeExerciseType(value);
  return type === EXERCISE_TYPE_STOPWATCH || type === EXERCISE_TYPE_COUNTDOWN;
}
