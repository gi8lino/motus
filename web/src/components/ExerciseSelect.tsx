import type { CatalogExercise } from "../types";
import { SelectDropdown } from "./SelectDropdown";

// ExerciseSelect renders a custom exercise picker with Core tags.
export function ExerciseSelect({
  catalog,
  value,
  onSelect,
  onClear,
  onAddNew,
}: {
  catalog: CatalogExercise[];
  value: { exerciseId?: string; name?: string };
  onSelect: (exercise: CatalogExercise) => void;
  onClear: () => void;
  onAddNew: () => void | Promise<void>;
}) {
  const selected = value.exerciseId
    ? catalog.find((item) => item.id === value.exerciseId)
    : undefined;
  const label = selected?.name || value.name || "";
  const tag = selected
    ? selected.isCore
      ? "core"
      : "user"
    : value.name
      ? "unlinked"
      : "";

  return (
    <SelectDropdown
      items={catalog.map((item) => ({ id: item.id, label: item.name }))}
      value={value.exerciseId || null}
      placeholder="Select exercise"
      onSelect={(item) => {
        const match = catalog.find((entry) => entry.id === item.id);
        if (match) onSelect(match);
      }}
      onClear={label ? onClear : undefined}
      renderRight={(item) => {
        const match = catalog.find((entry) => entry.id === item.id);
        if (!match) return null;
        return (
          <span className={`exercise-tag ${match.isCore ? "core" : "user"}`}>
            {match.isCore ? "Core" : "Personal"}
          </span>
        );
      }}
      renderSelectedRight={() =>
        tag ? (
          <span className={`exercise-tag ${tag}`}>
            {tag === "core" ? "Core" : tag === "user" ? "Personal" : "Unlinked"}
          </span>
        ) : null
      }
      addLabel="+ Add new exercise"
      onAddNew={onAddNew}
    />
  );
}
