import type { WorkoutStep } from "../../types";
import { UI_TEXT } from "../../utils/uiText";

export function WorkoutRepeatToggle({
  step,
  expanded,
  onToggle,
}: {
  step: WorkoutStep;
  expanded: boolean;
  onToggle: () => void;
}) {
  return (
    <>
      <button className="btn subtle" type="button" onClick={onToggle}>
        {expanded
          ? UI_TEXT.workouts.repeatOptions.hide
          : UI_TEXT.workouts.repeatOptions.show}
      </button>
      <span className="muted small">
        {step.repeatCount && step.repeatCount > 1
          ? `Repeats ${step.repeatCount}x`
          : UI_TEXT.workouts.repeatOptions.none}
      </span>
    </>
  );
}
