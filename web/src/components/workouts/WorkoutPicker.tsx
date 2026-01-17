import type { Workout } from "../../types";
import { UI_TEXT } from "../../utils/uiText";
import { SelectDropdown } from "../common/SelectDropdown";

// WorkoutPicker renders a custom workout picker styled like exercise inputs.
export function WorkoutPicker({
  workouts,
  value,
  onSelect,
  onClear,
}: {
  workouts: Workout[];
  value: string | null;
  onSelect: (workoutId: string) => void;
  onClear: () => void;
}) {
  return (
    <SelectDropdown
      items={workouts.map((w) => ({ id: w.id, label: w.name }))}
      value={value}
      placeholder={UI_TEXT.placeholders.selectWorkout}
      onSelect={(item) => onSelect(item.id)}
      onClear={onClear}
      renderRight={(item) => {
        const match = workouts.find((w) => w.id === item.id);
        if (!match) return null;
        return <span className="muted small">{match.steps.length} steps</span>;
      }}
    />
  );
}
