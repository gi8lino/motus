import assert from "node:assert/strict";
import test from "node:test";

import {
  runWorkoutSubmit,
  validateWorkoutDraft,
} from "../src/utils/workoutSubmit.ts";

test("workout draft validation reports actionable local errors", () => {
  assert.equal(validateWorkoutDraft("", []), "Enter a workout name.");
  assert.equal(validateWorkoutDraft("Plan", []), "Add at least one step.");
  assert.equal(
    validateWorkoutDraft("Plan", [{ type: "set", subsets: [] }]),
    "Step 1 needs at least one subset.",
  );
  assert.equal(
    validateWorkoutDraft("Plan", [
      { type: "set", subsets: [{ exercises: [] }] },
    ]),
    "Step 1, subset 1 needs at least one exercise.",
  );
  assert.equal(
    validateWorkoutDraft("Plan", [
      { type: "set", subsets: [{ exercises: [{}] }] },
    ]),
    null,
  );
  assert.equal(
    validateWorkoutDraft("Plan", [
      {
        type: "set",
        subsets: [
          {
            superset: true,
            exercises: [{ type: "rep" }, { type: "countdown" }],
          },
        ],
      },
    ]),
    "Step 1, subset 1: supersets support rep exercises only.",
  );
});

test("failed workout submit retains the backend error and skips success cleanup", async () => {
  let successCleanupRan = false;

  const result = await runWorkoutSubmit(
    async () => {
      throw new Error("step 2 requires at least one exercise");
    },
    () => {
      successCleanupRan = true;
    },
    "Unable to save workout",
  );

  assert.deepEqual(result, {
    ok: false,
    error: "step 2 requires at least one exercise",
  });
  assert.equal(successCleanupRan, false);
});

test("successful workout submit runs cleanup exactly once", async () => {
  let successCleanupCount = 0;

  const result = await runWorkoutSubmit(
    async () => {},
    () => {
      successCleanupCount += 1;
    },
    "Unable to save workout",
  );

  assert.deepEqual(result, { ok: true });
  assert.equal(successCleanupCount, 1);
});
