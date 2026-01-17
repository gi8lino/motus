import { useMemo } from "react";
import type { CatalogExercise } from "../../types";
import { UI_TEXT } from "../../utils/uiText";
import { SelectDropdown } from "../common/SelectDropdown";

type ExerciseValue = { exerciseId?: string; name?: string };

type ExerciseSelectProps = {
  catalog: CatalogExercise[];
  value: ExerciseValue;
  onSelect: (exercise: CatalogExercise) => void;
  onClear: () => void;
  onAddNew: () => void | Promise<void>;
};

// ExerciseSelect renders a custom exercise picker with Core tags.
export function ExerciseSelect({
  catalog,
  value,
  onSelect,
  onClear,
  onAddNew,
}: ExerciseSelectProps) {
  const byId = useMemo(() => new Map(catalog.map((e) => [e.id, e])), [catalog]);

  const selected = value.exerciseId ? byId.get(value.exerciseId) : undefined;
  const label = selected?.name || value.name || "";

  const tag: "" | "core" | "user" | "unlinked" = selected
    ? selected.isCore
      ? "core"
      : "user"
    : value.name
      ? "unlinked"
      : "";

  const items = useMemo(
    () => catalog.map((item) => ({ id: item.id, label: item.name })),
    [catalog],
  );

  return (
    <SelectDropdown
      items={items}
      value={value.exerciseId || null}
      placeholder={UI_TEXT.placeholders.selectExercise}
      onSelect={(item) => {
        const match = byId.get(item.id);
        if (match) onSelect(match);
      }}
      onClear={label ? onClear : undefined}
      renderRight={(item) => {
        const match = byId.get(item.id);
        if (!match) return null;
        return (
          <span className={`exercise-tag ${match.isCore ? "core" : "user"}`}>
            {match.isCore ? UI_TEXT.exercises.core : UI_TEXT.exercises.personal}
          </span>
        );
      }}
      renderSelectedRight={() =>
        tag ? (
          <span className={`exercise-tag ${tag}`}>
            {tag === "core"
              ? UI_TEXT.exercises.core
              : tag === "user"
                ? UI_TEXT.exercises.personal
                : UI_TEXT.exercises.unlinked}
          </span>
        ) : null
      }
      addLabel="+ Add new exercise"
      onAddNew={onAddNew}
    />
  );
}
