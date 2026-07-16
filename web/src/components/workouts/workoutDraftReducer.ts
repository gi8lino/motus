import type { SetStateAction } from "react";
import type { WorkoutStep } from "../../types";

export type WorkoutDraftAction = {
  type: "updateSteps";
  update: SetStateAction<WorkoutStep[]>;
};

export function workoutStepsReducer(steps: WorkoutStep[], action: WorkoutDraftAction) {
  switch (action.type) {
    case "updateSteps":
      return typeof action.update === "function" ? action.update(steps) : action.update;
  }
}
