import type { SetStateAction } from "react";
import type { Exercise, Workout, WorkoutStep } from "../../types";

export function cloneWorkoutStep(step: WorkoutStep): WorkoutStep {
  return {
    ...step,
    id: undefined,
    subsets: step.subsets?.map((subset) => ({
      ...subset,
      id: undefined,
      exercises: subset.exercises?.map((exercise) => ({ ...exercise })),
    })),
    exercises: step.exercises?.map((exercise) => ({ ...exercise })),
  };
}

export function duplicateWorkoutDraft(workout: Workout) {
  return {
    name: `${workout.name} (copy)`,
    steps: workout.steps.map(cloneWorkoutStep),
  };
}

export type WorkoutDraftAction =
  | { type: "updateSteps"; update: SetStateAction<WorkoutStep[]> }
  | { type: "addStep"; step: WorkoutStep }
  | { type: "updateStep"; index: number; patch: Partial<WorkoutStep> }
  | { type: "removeStep"; index: number }
  | { type: "duplicateStep"; index: number }
  | { type: "moveStep"; index: number; delta: number }
  | {
      type: "updateExercise";
      stepIndex: number;
      subsetIndex: number;
      exerciseIndex: number;
      patch: Partial<Exercise>;
    }
  | {
      type: "moveExercise";
      stepIndex: number;
      subsetIndex: number;
      from: number;
      to: number;
    }
  | {
      type: "addExercise";
      stepIndex: number;
      subsetIndex: number;
      exercise: Exercise;
    }
  | {
      type: "removeExercise";
      stepIndex: number;
      subsetIndex: number;
      exerciseIndex: number;
    };

function updateExercises(
  steps: WorkoutStep[],
  stepIndex: number,
  subsetIndex: number,
  update: (items: Exercise[]) => Exercise[],
) {
  return steps.map((step, index) => {
    if (index !== stepIndex) return step;
    const subsets = [...(step.subsets || [])];
    const subset = subsets[subsetIndex];
    if (!subset) return step;
    subsets[subsetIndex] = {
      ...subset,
      exercises: update([...(subset.exercises || [])]),
    };
    return { ...step, subsets };
  });
}

export function workoutStepsReducer(
  steps: WorkoutStep[],
  action: WorkoutDraftAction,
): WorkoutStep[] {
  switch (action.type) {
    case "updateSteps":
      return typeof action.update === "function"
        ? action.update(steps)
        : action.update;
    case "addStep":
      return [...steps, action.step];
    case "updateStep":
      return steps.map((step, index) =>
        index === action.index ? { ...step, ...action.patch } : step,
      );
    case "removeStep":
      return steps.filter((_, index) => index !== action.index);
    case "duplicateStep": {
      const source = steps[action.index];
      if (!source) return steps;
      const next = [...steps];
      next.splice(action.index + 1, 0, cloneWorkoutStep(source));
      return next;
    }
    case "moveStep": {
      const target = action.index + action.delta;
      if (target < 0 || target >= steps.length) return steps;
      const next = [...steps];
      const [item] = next.splice(action.index, 1);
      next.splice(target, 0, item);
      return next;
    }
    case "updateExercise":
      return updateExercises(
        steps,
        action.stepIndex,
        action.subsetIndex,
        (items) =>
          items.map((item, index) =>
            index === action.exerciseIndex
              ? { ...item, ...action.patch }
              : item,
          ),
      );
    case "moveExercise":
      return updateExercises(
        steps,
        action.stepIndex,
        action.subsetIndex,
        (items) => {
          const [item] = items.splice(action.from, 1);
          items.splice(action.to, 0, item);
          return items;
        },
      );
    case "addExercise":
      return updateExercises(
        steps,
        action.stepIndex,
        action.subsetIndex,
        (items) => [...items, action.exercise],
      );
    case "removeExercise":
      return updateExercises(
        steps,
        action.stepIndex,
        action.subsetIndex,
        (items) => items.filter((_, index) => index !== action.exerciseIndex),
      );
  }
}
