import test from "node:test";
import assert from "node:assert/strict";

import {
  formatExerciseMetric,
  formatRoundValue,
  formatExerciseSide,
  formatStepValue,
} from "../src/utils/trainingCard.ts";
import {
  duplicateWorkoutDraft,
  workoutStepsReducer,
} from "../src/components/workouts/workoutDraftReducer.ts";

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

test("training card helpers cover metric, counters, and rounds", () => {
  assert.equal(
    formatExerciseMetric({ name: "Row", reps: "8", weight: "20kg" }),
    "8 reps · 20kg",
  );
  assert.equal(
    formatExerciseMetric({
      name: "Single-arm row",
      reps: "8",
      side: "left",
    }),
    "8 reps",
  );
  assert.equal(
    formatExerciseMetric({
      name: "Single-arm row",
      duration: "20s",
      side: "right",
    }),
    "20s",
  );
  assert.equal(
    formatExerciseSide({ name: "Single-arm row", side: "left" }),
    "Left",
  );
  assert.equal(formatStepValue(2, 5), "2/5");
  assert.equal(
    formatRoundValue({
      name: "Round",
      type: "set",
      elapsedMillis: 0,
      completed: false,
      current: true,
      running: false,
      loopIndex: 2,
      loopTotal: 3,
    }),
    "2/3",
  );
});

test("duplicate workout and step create independent nested drafts", () => {
  const workout = {
    id: "w1",
    userId: "u1",
    name: "Strength",
    steps: [
      {
        id: "step-1",
        name: "Set",
        type: "set" as const,
        subsets: [
          {
            id: "subset-1",
            name: "Main",
            exercises: [{ name: "Squat", reps: "5" }],
          },
        ],
      },
    ],
  };
  const draft = duplicateWorkoutDraft(workout);
  assert.equal(draft.name, "Strength (copy)");
  assert.equal(draft.steps[0].id, undefined);
  assert.equal(draft.steps[0].subsets?.[0].id, undefined);

  const duplicated = workoutStepsReducer(workout.steps, {
    type: "duplicateStep",
    index: 0,
  });
  assert.equal(duplicated.length, 2);
  duplicated[1].subsets![0].exercises![0].reps = "10";
  assert.equal(duplicated[0].subsets![0].exercises![0].reps, "5");
});
