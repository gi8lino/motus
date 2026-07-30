import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";

type FixtureExercise = {
  type?: string;
  duration?: string;
  side?: string;
};

type FixtureStep = {
  type: string;
  estimatedSeconds?: number;
  repeatCount?: number;
  repeatRestSeconds?: number;
  pauseOptions?: { autoAdvance?: boolean };
  subsets?: Array<{
    superset?: boolean;
    exercises?: FixtureExercise[];
  }>;
};

test("test workout covers training modes and keeps timers short", async () => {
  const raw = await readFile(
    new URL("../public/test-workout.json", import.meta.url),
    "utf8",
  );
  const fixture = JSON.parse(raw) as { name: string; steps: FixtureStep[] };
  const exercises = fixture.steps.flatMap((step) =>
    (step.subsets || []).flatMap((subset) => subset.exercises || []),
  );
  const durations = exercises
    .map((exercise) => exercise.duration)
    .filter((duration): duration is string => Boolean(duration))
    .map((duration) => Number.parseInt(duration, 10));

  assert.match(fixture.name, /Test Workout/);
  assert.ok(exercises.some((exercise) => exercise.type === "rep"));
  assert.ok(exercises.some((exercise) => exercise.type === "countdown"));
  assert.ok(exercises.some((exercise) => exercise.type === "stopwatch"));
  assert.ok(exercises.some((exercise) => exercise.side === "left"));
  assert.ok(exercises.some((exercise) => exercise.side === "right"));
  assert.ok(
    fixture.steps.some((step) =>
      step.subsets?.some((subset) => subset.superset),
    ),
  );
  assert.ok(fixture.steps.some((step) => (step.repeatCount || 0) > 1));
  assert.ok(
    fixture.steps.some(
      (step) => step.type === "pause" && step.pauseOptions?.autoAdvance,
    ),
  );
  assert.ok(
    fixture.steps.some(
      (step) => step.type === "pause" && !step.pauseOptions?.autoAdvance,
    ),
  );

  const timerSeconds = [
    ...durations,
    ...fixture.steps.map((step) => step.estimatedSeconds || 0),
    ...fixture.steps.map((step) => step.repeatRestSeconds || 0),
  ].filter(Boolean);
  assert.ok(timerSeconds.length > 0);
  assert.ok(timerSeconds.every((seconds) => seconds <= 6));
});
