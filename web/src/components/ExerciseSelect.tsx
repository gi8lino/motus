import { useEffect, useRef, useState } from "react";
import type { CatalogExercise } from "../types";

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
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement | null>(null);
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

  useEffect(() => {
    const handleClick = (event: MouseEvent) => {
      if (!rootRef.current) return;
      if (rootRef.current.contains(event.target as Node)) return;
      setOpen(false);
    };
    if (open) {
      document.addEventListener("mousedown", handleClick);
    }
    return () => document.removeEventListener("mousedown", handleClick);
  }, [open]);

  return (
    <div className="exercise-select" ref={rootRef}>
      <button
        className="exercise-select-trigger"
        type="button"
        onClick={() => setOpen((prev) => !prev)}
      >
        <span className={label ? "" : "muted"}>
          {label || "Select exercise"}
        </span>
        {tag && (
          <span className={`exercise-tag ${tag}`}>
            {tag === "core" ? "Core" : tag === "user" ? "Personal" : "Unlinked"}
          </span>
        )}
        <span className="chevron">{open ? "▼" : "▶"}</span>
      </button>
      {open && (
        <div className="exercise-select-menu">
          {label && (
            <button
              className="exercise-select-option muted"
              type="button"
              onClick={() => {
                onClear();
                setOpen(false);
              }}
            >
              Clear selection
            </button>
          )}
          {catalog.map((item) => (
            <button
              key={item.id}
              className="exercise-select-option"
              type="button"
              onClick={() => {
                onSelect(item);
                setOpen(false);
              }}
            >
              <span>{item.name}</span>
              <span className={`exercise-tag ${item.isCore ? "core" : "user"}`}>
                {item.isCore ? "Core" : "Personal"}
              </span>
            </button>
          ))}
          <button
            className="exercise-select-option add"
            type="button"
            onClick={async () => {
              await onAddNew();
              setOpen(false);
            }}
          >
            + Add new exercise
          </button>
        </div>
      )}
    </div>
  );
}
