type WorkoutSubmitResult = { ok: true } | { ok: false; error: string };

export async function runWorkoutSubmit(
  save: () => Promise<void>,
  onSuccess: () => void,
  fallbackError: string,
): Promise<WorkoutSubmitResult> {
  try {
    await save();
    onSuccess();
    return { ok: true };
  } catch (error) {
    const message =
      error instanceof Error && error.message
        ? error.message
        : typeof error === "string" && error
          ? error
          : fallbackError;
    return { ok: false, error: message };
  }
}
