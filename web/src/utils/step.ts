export const STEP_TYPE_SET = "set";
export const STEP_TYPE_PAUSE = "pause";

export type StepType = typeof STEP_TYPE_SET | typeof STEP_TYPE_PAUSE;

export function normalizeStepType(value?: string): StepType {
  const token = (value || "").trim().toLowerCase();
  if (token === STEP_TYPE_PAUSE) return STEP_TYPE_PAUSE;
  return STEP_TYPE_SET;
}

export function isPauseStepType(value?: string): boolean {
  return normalizeStepType(value) === STEP_TYPE_PAUSE;
}

export function isSetStepType(value?: string): boolean {
  return normalizeStepType(value) === STEP_TYPE_SET;
}
