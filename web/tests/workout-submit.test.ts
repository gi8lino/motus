import assert from "node:assert/strict";
import test from "node:test";

import { runWorkoutSubmit } from "../src/utils/workoutSubmit.ts";

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
