// MESSAGES centralizes user-facing fallback text for errors.
export const MESSAGES = {
  loadFailed: "Unable to load data",
  authFailed: "Unable to authenticate user",
  loginFailed: "Invalid login",
  configFailed: "Unable to load configuration",
  saveWorkoutFailed: "Unable to save workout",
  updateWorkoutFailed: "Unable to update workout",
  deleteWorkoutFailed: "Unable to delete workout",
  shareWorkoutFailed: "Unable to share workout",
  copyTemplateFailed: "Unable to copy to clipboard",
  applyTemplateFailed: "Unable to apply template",
  startTrainingFailed: "Unable to start training",
  saveTrainingFailed: "Unable to save training.",
  createExerciseFailed: "Unable to create exercise",
  createCoreExerciseFailed: "Unable to create core exercise",
  renameExerciseFailed: "Unable to rename exercise",
  deleteExerciseFailed: "Unable to delete exercise",
  exportWorkoutFailed: "Unable to export workout",
  importWorkoutFailed: "Unable to import workout",
  updatePasswordFailed: "Unable to update password",
  updateRoleFailed: "Unable to update role",
  backfillExercisesFailed: "Unable to backfill exercises",
  updateNameFailed: "Unable to update name",
  logTrainingFailed: "Unable to log training",
} as const;

// toErrorMessage normalizes error values into display-friendly text.
export function toErrorMessage(err: unknown, fallback: string): string {
  if (err instanceof Error && err.message) return err.message;
  if (typeof err === "string" && err) return err;
  return fallback;
}
