import type { Exercise } from "../types";
import { formatExerciseLine, formatMillis } from "./format";
import { normalizeTimestamp } from "./time";

// SummaryStep describes a step entry for AI summaries.
export type SummaryStep = {
  name: string;
  type?: string;
  estimatedSeconds?: number;
  elapsedMillis?: number;
  exercises?: Exercise[];
};

// SummarySource is the minimal input required to build a summary.
export type SummarySource = {
  workoutName?: string;
  workoutId?: string;
  userId?: string;
  startedAt?: string | null;
  completedAt?: string | null;
  steps: SummaryStep[];
};

// buildSummary renders a copyable summary for AI tools.
export function buildSummary(source: SummarySource): string {
  const lines = source.steps
    .map((s, idx) => {
      // Build target/actual labels per step.
      const actual = s.elapsedMillis
        ? `actual ${formatMillis(s.elapsedMillis)}`
        : "";
      const target = s.estimatedSeconds ? `target ${s.estimatedSeconds}s` : "";
      const label = [target, actual].filter(Boolean).join(", ");
      const exercises =
        s.exercises && s.exercises.length
          ? ` — ${s.exercises
              .map((ex: Exercise) => formatExerciseLine(ex))
              .filter(Boolean)
              .join(" | ")}`
          : "";
      return `${idx + 1}. ${s.name}${label ? ` (${label})` : ""}${exercises}`;
    })
    .join("\n");
  const startedAt = normalizeTimestamp(source.startedAt) || "n/a";
  const completedAt = normalizeTimestamp(source.completedAt) || "n/a";
  return `Workout: ${source.workoutName || source.workoutId || "n/a"}
User: ${source.userId || "n/a"}
Started: ${startedAt}
Finished: ${completedAt}
Steps:
${lines || "No steps available."}`;
}
