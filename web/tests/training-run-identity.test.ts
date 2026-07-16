import test from "node:test";
import assert from "node:assert/strict";
import {
  buildStepRunKey,
  isSameStepRun,
} from "../src/hooks/trainingTimer/runIdentity.ts";

const step = {
  id: "step",
  name: "Work",
  type: "set" as const,
  elapsedMillis: 0,
  completed: false,
  current: true,
  running: true,
  estimatedSeconds: 30,
};
const state = {
  trainingId: "training",
  workoutId: "workout",
  userId: "user",
  currentIndex: 0,
  running: true,
  runningSince: 100,
  done: false,
  steps: [step],
};

test("step run identity changes with timer anchors", () => {
  assert.equal(buildStepRunKey(state, step, 30), "training:0:step:100:30");
  assert.equal(isSameStepRun(state, state, step, step, 30), true);
  assert.equal(
    isSameStepRun({ ...state, runningSince: 200 }, state, step, step, 30),
    false,
  );
});
