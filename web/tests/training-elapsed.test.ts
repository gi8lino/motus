import test from "node:test";
import assert from "node:assert/strict";
import { getTotalElapsedMillis } from "../src/utils/trainingElapsed.ts";
import type { TrainingState } from "../src/types.ts";

test("total elapsed uses the live value for the current step", () => {
  const training = {
    currentIndex: 1,
    steps: [
      { elapsedMillis: 10_000 },
      { elapsedMillis: 2_000 },
      { elapsedMillis: 0 },
    ],
  } as TrainingState;

  assert.equal(getTotalElapsedMillis(training, 4_500), 14_500);
});
