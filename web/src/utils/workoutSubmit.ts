type WorkoutSubmitResult = { ok: true } | { ok: false; error: string };

export function validateWorkoutDraft(
  name: string,
  steps: Array<{
    type: string;
    subsets?: Array<{ exercises?: unknown[] }>;
  }>,
): string | null {
  if (!name.trim()) return "Enter a workout name.";
  if (!steps.length) return "Add at least one step.";

  for (const [stepIndex, step] of steps.entries()) {
    if (step.type !== "set") continue;
    if (!step.subsets?.length) {
      return `Step ${stepIndex + 1} needs at least one subset.`;
    }
    const emptySubset = step.subsets.findIndex(
      (subset) => !subset.exercises?.length,
    );
    if (emptySubset >= 0) {
      return `Step ${stepIndex + 1}, subset ${emptySubset + 1} needs at least one exercise.`;
    }
  }
  return null;
}

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
