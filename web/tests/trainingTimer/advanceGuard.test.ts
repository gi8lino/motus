import assert from "node:assert/strict";
import test from "node:test";

import {
  claimAdvanceTransition,
  syncAdvanceTransition,
} from "../../src/hooks/trainingTimer/advanceGuard.ts";
import type { TrainingState } from "../../src/types.ts";

function makeTrainingState(overrides?: Partial<TrainingState>): TrainingState {
  return {
    trainingId: "training-1",
    workoutId: "workout-1",
    workoutName: "Workout",
    userId: "user-1",
    currentIndex: 0,
    running: true,
    runningSince: 1000,
    done: false,
    startedAt: "2026-05-15T10:00:00.000Z",
    completedAt: null,
    logged: false,
    steps: [
      {
        id: "step-1",
        type: "set",
        name: "Step 1",
        elapsedMillis: 0,
        completed: false,
        current: true,
        running: true,
      },
      {
        id: "step-2",
        type: "set",
        name: "Step 2",
        elapsedMillis: 0,
        completed: false,
        current: false,
        running: false,
      },
    ],
    ...overrides,
  };
}

test("claimAdvanceTransition blocks duplicate consumption for the same step run", () => {
  const guardRef = { current: null as string | null };
  const training = makeTrainingState();

  assert.equal(claimAdvanceTransition(guardRef, training), true);
  assert.equal(claimAdvanceTransition(guardRef, training), false);
});

test("syncAdvanceTransition reopens the guard after the training moves on", () => {
  const guardRef = { current: null as string | null };
  const firstStep = makeTrainingState();

  assert.equal(claimAdvanceTransition(guardRef, firstStep), true);

  const secondStep = makeTrainingState({
    currentIndex: 1,
    runningSince: 2000,
    steps: firstStep.steps.map((step, index) => ({
      ...step,
      current: index === 1,
      running: index === 1,
      completed: index < 1,
    })),
  });

  syncAdvanceTransition(guardRef, secondStep);

  assert.equal(claimAdvanceTransition(guardRef, secondStep), true);
});
