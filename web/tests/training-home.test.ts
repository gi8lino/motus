import test from "node:test";
import assert from "node:assert/strict";
import { selectQuickWorkouts } from "../src/utils/trainingHome.ts";

const workouts = ["a", "b", "c"].map((id) => ({
  id,
  userId: "user",
  name: id.toUpperCase(),
  steps: [],
  favorite: id === "c",
}));

test("quick workouts put favorites before unique recent workouts", () => {
  const result = selectQuickWorkouts(workouts, [
    {
      id: "1",
      trainingId: "1",
      workoutId: "a",
      userId: "user",
      completedAt: "2026-01-01",
    },
    {
      id: "2",
      trainingId: "2",
      workoutId: "b",
      userId: "user",
      completedAt: "2026-02-01",
    },
  ]);
  assert.deepEqual(
    result.map((workout) => workout.id),
    ["c", "b", "a"],
  );
});
