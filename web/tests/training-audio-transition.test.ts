import test from "node:test";
import assert from "node:assert/strict";
import {
  didStepChange,
  getStepSoundTargetSeconds,
} from "../src/hooks/trainingAudio/transitions.ts";

test("pause sound targets the end of its configured duration", () => {
  assert.equal(
    getStepSoundTargetSeconds(
      {
        type: "pause",
        name: "Rest",
        estimatedSeconds: 30,
        elapsedMillis: 0,
        completed: false,
        current: true,
        running: true,
      },
      0,
    ),
    30,
  );
});

test("step transitions are distinguished from pausing the same step", () => {
  assert.equal(didStepChange("pause-1", "exercise-2"), true);
  assert.equal(didStepChange("pause-1", "pause-1"), false);
});
