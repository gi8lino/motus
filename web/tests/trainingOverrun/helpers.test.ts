import assert from "node:assert/strict";
import test from "node:test";

import {
  buildOverrunKey,
  clearTimer,
  getEffectiveThresholdMs,
  getOverrunThresholdMs,
} from "../../src/hooks/trainingOverrun/helpers.ts";

test("buildOverrunKey returns null when training is missing", () => {
  assert.equal(buildOverrunKey(null, null), null);
});

test("buildOverrunKey returns training-step key", () => {
  const training = {
    trainingId: "t1",
    workoutId: "w1",
    userId: "u1",
    currentIndex: 2,
    running: true,
    done: false,
    steps: [],
  };
  const currentStep = { id: "s1" };

  assert.equal(buildOverrunKey(training, currentStep), "t1:2:s1");
});

test("getOverrunThresholdMs adds grace to estimated step duration", () => {
  assert.equal(getOverrunThresholdMs(undefined), 0);
  assert.equal(getOverrunThresholdMs(0), 0);
  assert.equal(getOverrunThresholdMs(15), 45_000);
});

test("getEffectiveThresholdMs uses postponed threshold when larger", () => {
  assert.equal(getEffectiveThresholdMs(10_000, null), 10_000);
  assert.equal(getEffectiveThresholdMs(10_000, 8_000), 10_000);
  assert.equal(getEffectiveThresholdMs(10_000, 12_500), 12_500);
});

test("clearTimer clears timeout reference", async () => {
  let fired = false;
  const timerRef = {
    current: setTimeout(() => {
      fired = true;
    }, 20) as unknown as number,
  };

  clearTimer(timerRef);
  assert.equal(timerRef.current, null);

  await new Promise((resolve) => setTimeout(resolve, 40));
  assert.equal(fired, false);
});
