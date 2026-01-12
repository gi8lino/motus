export const STEP_TYPE_SET = "set";
export const STEP_TYPE_PAUSE = "pause";

export type StepType = typeof STEP_TYPE_SET | typeof STEP_TYPE_PAUSE;

// normalizeStepType coerces raw values into a known step type.
export function normalizeStepType(value?: string): StepType {
  const token = (value || "").trim().toLowerCase();
  if (token === STEP_TYPE_PAUSE) return STEP_TYPE_PAUSE;
  return STEP_TYPE_SET;
}

// isPauseStepType reports whether a value represents a pause step.
export function isPauseStepType(value?: string): boolean {
  return normalizeStepType(value) === STEP_TYPE_PAUSE;
}

// isSetStepType reports whether a value represents a set step.
export function isSetStepType(value?: string): boolean {
  return normalizeStepType(value) === STEP_TYPE_SET;
}
