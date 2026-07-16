import test from "node:test";
import assert from "node:assert/strict";

import {
  formatExerciseMetric,
  formatRoundLabel,
  formatStepValue,
  getAdaptiveTitleSize,
} from "../src/utils/trainingCard.ts";
import { workoutStepsReducer } from "../src/components/workouts/workoutDraftReducer.ts";

test("workout steps reducer supports replacement and functional updates", () => {
  const initial = [{ name: "One", type: "set" as const }];
  const replaced = workoutStepsReducer(initial, {
    type: "updateSteps",
    update: [{ name: "Two", type: "set" }],
  });
  assert.equal(replaced[0].name, "Two");
  const appended = workoutStepsReducer(replaced, {
    type: "updateSteps",
    update: (steps) => [...steps, { name: "Pause", type: "pause" }],
  });
  assert.equal(appended.length, 2);
});

test("training card helpers cover metric, counters, rounds, and title sizing", () => {
  assert.equal(
    formatExerciseMetric({ name: "Row", reps: "8", weight: "20kg" }),
    "8 reps · 20kg",
  );
  assert.equal(formatStepValue(2, 5), "2/5");
  assert.equal(
    formatRoundLabel({
      name: "Round",
      type: "set",
      elapsedMillis: 0,
      completed: false,
      current: true,
      running: false,
      loopIndex: 2,
      loopTotal: 3,
    }),
    "Round 2/3",
  );
  assert.equal(
    getAdaptiveTitleSize("A very long exercise title indeed", {
      short: "s",
      medium: "m",
      long: "l",
      xlong: "xl",
    }),
    "xl",
  );
});
