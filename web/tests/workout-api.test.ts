import assert from "node:assert/strict";
import test from "node:test";

import { workoutWriteSteps } from "../src/utils/workoutWrite.ts";

test("workout write payload strips editor ids and keeps exercise side", () => {
  const [step] = workoutWriteSteps([
    {
      id: "local-step",
      order: 4,
      type: "set",
      name: "Strength",
      subsets: [
        {
          id: "local-subset",
          name: "Main",
          exercises: [{ name: "Kettlebell row", reps: "8", side: "left" }],
        },
      ],
    },
  ]);

  assert.equal("id" in step, false);
  assert.equal("order" in step, false);
  assert.equal("id" in step.subsets[0], false);
  assert.equal(step.subsets[0].exercises[0].side, "left");
});
