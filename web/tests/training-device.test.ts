import test from "node:test";
import assert from "node:assert/strict";
import { shouldVibrate } from "../src/hooks/useTrainingDevice.ts";

test("vibration requires both preference and browser support", () => {
  assert.equal(shouldVibrate(true, true), true);
  assert.equal(shouldVibrate(false, true), false);
  assert.equal(shouldVibrate(true, false), false);
});
