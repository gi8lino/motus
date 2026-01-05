import { useEffect, useMemo, useRef, useState, FormEvent } from "react";
import type {
  CatalogExercise,
  Exercise,
  SoundOption,
  Workout,
  WorkoutStep,
} from "../types";

const DEFAULT_WORKOUT_NAME = "Push Day";

// WorkoutForm lets you create or edit a workout with steps/exercises.
export function WorkoutForm({
  onSave,
  onUpdate,
  editingWorkout,
  sounds,
  userId,
  exerciseCatalog,
  onCreateExercise,
  onClose,
  promptUser,
  notifyUser,
  onDirtyChange,
  onToast,
}: {
  onSave: (payload: { name: string; steps: WorkoutStep[] }) => Promise<void>;
  onUpdate?: (payload: {
    id: string;
    name: string;
    steps: WorkoutStep[];
  }) => Promise<void>;
  editingWorkout?: Workout | null;
  sounds: SoundOption[];
  userId: string | null;
  exerciseCatalog?: CatalogExercise[];
  onCreateExercise: (name: string) => Promise<CatalogExercise>;
  onClose?: () => void;
  promptUser: (
    message: string,
    defaultValue?: string,
  ) => Promise<string | null>;
  notifyUser: (message: string) => Promise<void>;
  onDirtyChange?: (dirty: boolean) => void;
  onToast?: (message: string) => void;
}) {
  const [name, setName] = useState(DEFAULT_WORKOUT_NAME);
  const [steps, setSteps] = useState<WorkoutStep[]>([]);
  const [editingId, setEditingId] = useState<string | null>(null);
  const dragIndex = useRef<number | null>(null);
  const dragExerciseRef = useRef<{ stepIdx: number; idx: number } | null>(null);
  const [expandedSteps, setExpandedSteps] = useState<Set<number>>(new Set([0]));
  const [dirty, setDirty] = useState(false);
  const catalog = exerciseCatalog || [];
  const catalogById = useMemo(
    () => new Map(catalog.map((entry) => [entry.id, entry])),
    [catalog],
  );
  const catalogByName = useMemo(
    () => new Map(catalog.map((entry) => [entry.name.toLowerCase(), entry])),
    [catalog],
  );

  useEffect(() => {
    if (!editingWorkout) {
      setEditingId(null);
      setName(DEFAULT_WORKOUT_NAME);
      setSteps([]);
      setExpandedSteps(new Set([0]));
      setDirty(false);
      onDirtyChange?.(false);
      return;
    }

    setEditingId(editingWorkout.id);
    setName(editingWorkout.name);
    setSteps(
      (editingWorkout.steps || []).map((s) => ({
        ...s,
        pauseOptions:
          s.pauseOptions ||
          (s.weight === "__auto__" ? { autoAdvance: true } : undefined),
        duration:
          s.duration || (s.estimatedSeconds ? `${s.estimatedSeconds}s` : ""),
        exercises:
          s.exercises && s.exercises.length
            ? s.exercises
            : s.type === "set"
              ? [
                  {
                    name: s.exercises?.[0]?.name || s.name || "",
                    amount: s.exercises?.[0]?.amount || "",
                    weight: s.exercises?.[0]?.weight || "",
                    exerciseId: s.exercises?.[0]?.exerciseId,
                  },
                ]
              : [],
      })),
    );
    setDirty(false);
    onDirtyChange?.(false);
  }, [editingWorkout, onDirtyChange]);

  useEffect(() => {
    if (!catalog.length) return;
    setSteps((prev) =>
      prev.map((step) => ({
        ...step,
        exercises: (step.exercises || []).map((ex) => {
          if (ex.exerciseId || !ex.name) return ex;
          const resolved = catalogByName.get(ex.name.toLowerCase());
          if (!resolved) return ex;
          return { ...ex, exerciseId: resolved.id };
        }),
      })),
    );
  }, [catalogByName, catalog.length]);

  // addStep appends a new default step and expands it for editing.
  const addStep = () =>
    setSteps((prev) => {
      const newStep: WorkoutStep = {
        type: "set",
        name: `Step ${prev.length + 1}`,
        duration: "1m",
      };
      const next = [...prev, newStep];
      setExpandedSteps((exp) => new Set(exp).add(next.length - 1));
      setDirty(true);
      onDirtyChange?.(true);
      return next;
    });

  // updateStep merges a partial update into a step at an index.
  const updateStep = (idx: number, patch: Partial<WorkoutStep>) => {
    setSteps((prev) =>
      prev.map((step, i) => (i === idx ? { ...step, ...patch } : step)),
    );
    setDirty(true);
    onDirtyChange?.(true);
  };

  // removeStep deletes a step by index and collapses it.
  const removeStep = (idx: number) => {
    setSteps((prev) => prev.filter((_, i) => i !== idx));
    setExpandedSteps((prev) => {
      const next = new Set(prev);
      next.delete(idx);
      return next;
    });
    setDirty(true);
    onDirtyChange?.(true);
  };

  // moveStep reorders a step by a delta offset.
  const moveStep = (index: number, delta: number) => {
    setSteps((prev) => {
      const next = [...prev];
      const target = index + delta;
      if (target < 0 || target >= next.length) return prev;
      const [item] = next.splice(index, 1);
      next.splice(target, 0, item);
      return next;
    });
    setDirty(true);
    onDirtyChange?.(true);
  };

  // moveExercise reorders an exercise inside a step.
  const moveExercise = (stepIdx: number, from: number, to: number) => {
    setSteps((prev) =>
      prev.map((step, idx) => {
        if (idx !== stepIdx) return step;
        const updated = [...(step.exercises || [])];
        const [item] = updated.splice(from, 1);
        updated.splice(to, 0, item);
        return { ...step, exercises: updated };
      }),
    );
    setDirty(true);
    onDirtyChange?.(true);
  };

  // updateExercise mutates an exercise at an index within a step.
  const updateExercise = (
    stepIdx: number,
    exIdx: number,
    patch: Partial<Exercise>,
  ) => {
    setSteps((prev) =>
      prev.map((step, idx) => {
        if (idx !== stepIdx) return step;
        const updated = [...(step.exercises || [])];
        updated[exIdx] = { ...updated[exIdx], ...patch };
        return { ...step, exercises: updated };
      }),
    );
    setDirty(true);
    onDirtyChange?.(true);
  };

  // addExercise appends a blank exercise row to a step.
  const addExercise = (stepIdx: number) => {
    setSteps((prev) =>
      prev.map((step, idx) =>
        idx === stepIdx
          ? {
              ...step,
              exercises: [
                ...(step.exercises || []),
                { name: "", amount: "", weight: "", exerciseId: "" },
              ],
            }
          : step,
      ),
    );
    setDirty(true);
    onDirtyChange?.(true);
  };

  // removeExercise deletes an exercise row.
  const removeExercise = (stepIdx: number, exIdx: number) => {
    setSteps((prev) =>
      prev.map((step, idx) => {
        if (idx !== stepIdx) return step;
        const updated = [...(step.exercises || [])];
        updated.splice(exIdx, 1);
        return { ...step, exercises: updated };
      }),
    );
    setDirty(true);
    onDirtyChange?.(true);
  };

  const durationLabel = useMemo(
    () => ({
      set: "Target time (optional, e.g. 45s)",
      pause: "Duration (Go style, e.g. 45s, 1m30s)",
      timed: "",
    }),
    [],
  );

  // renderStandardExercises shows normal set inputs with drag handles.
  function renderStandardExercises(idx: number, step: WorkoutStep) {
    return (
      <div className="stack">
        <div className="label">Exercises</div>
        {(step.exercises || []).map((ex: Exercise, exIdx) => (
          <div
            key={exIdx}
            className="exercise-row"
            draggable
            onDragStart={(e) => {
              e.stopPropagation();
              dragExerciseRef.current = { stepIdx: idx, idx: exIdx };
              e.dataTransfer.effectAllowed = "move";
            }}
            onDragOver={(e) => {
              e.preventDefault();
              const dragData = dragExerciseRef.current;
              if (!dragData || dragData.stepIdx !== idx) return;
              if (dragData.idx === exIdx) return;
              moveExercise(idx, dragData.idx, exIdx);
              dragExerciseRef.current = { stepIdx: idx, idx: exIdx };
            }}
            onDragEnd={() => {
              dragExerciseRef.current = null;
            }}
          >
            <div className="field">
              <label>Exercise</label>
              <select
                value={ex.exerciseId || (ex.name ? `name:${ex.name}` : "")}
                onChange={async (e) => {
                  const val = e.target.value;
                  if (val === "__add_new") {
                    const newName = await promptUser("Exercise name");
                    if (!newName || !newName.trim()) return;
                    try {
                      const created = await onCreateExercise(newName.trim());
                      updateExercise(idx, exIdx, {
                        name: created.name,
                        exerciseId: created.id,
                      });
                    } catch (err: any) {
                      await notifyUser(
                        err.message || "Unable to create exercise",
                      );
                    }
                    return;
                  }
                  if (!val) {
                    updateExercise(idx, exIdx, {
                      name: "",
                      exerciseId: "",
                    });
                    return;
                  }
                  if (val.startsWith("name:")) return;
                  const selected = catalogById.get(val);
                  if (!selected) return;
                  updateExercise(idx, exIdx, {
                    name: selected.name,
                    exerciseId: selected.id,
                  });
                }}
              >
                <option value="">Select exercise</option>
                {ex.exerciseId && !catalogById.has(ex.exerciseId) && (
                  <option value={ex.exerciseId} disabled>
                    {ex.name || "Unknown exercise"}
                  </option>
                )}
                {!ex.exerciseId && ex.name && (
                  <option value={`name:${ex.name}`} disabled>
                    {ex.name} (unlinked)
                  </option>
                )}
                {catalog.map((c) => (
                  <option
                    key={c.id}
                    value={c.id}
                    className={`exercise-option ${c.isCore ? "core" : "user"}`}
                  >
                    {c.isCore ? `${c.name} ★` : c.name}
                  </option>
                ))}
                <option value="__add_new">+ Add new exercise</option>
              </select>
            </div>
            <div className="field narrow">
              <label>Amount / Reps</label>
              <input
                value={ex.amount || ""}
                onChange={(e) =>
                  updateExercise(idx, exIdx, { amount: e.target.value })
                }
                placeholder="12 reps"
              />
            </div>
            <div className="field narrow">
              <label>Weight</label>
              <input
                value={ex.weight || ""}
                onChange={(e) =>
                  updateExercise(idx, exIdx, { weight: e.target.value })
                }
                placeholder="50kg"
              />
            </div>
            <button
              className="btn icon"
              type="button"
              onClick={() => removeExercise(idx, exIdx)}
              title="Remove exercise"
            >
              ×
            </button>
          </div>
        ))}
        <button
          className="btn outline"
          type="button"
          onClick={() => addExercise(idx)}
        >
          Add Exercise
        </button>
      </div>
    );
  }

  // renderTimedExercises shows duration/transition inputs for timed sets.
  function renderTimedExercises(idx: number, step: WorkoutStep) {
    return (
      <div className="stack">
        <div className="label">Timed Exercises</div>
        {(step.exercises || []).map((ex: Exercise, exIdx) => (
          <div
            key={exIdx}
            className="exercise-row"
            draggable
            onDragStart={(e) => {
              e.stopPropagation();
              dragExerciseRef.current = { stepIdx: idx, idx: exIdx };
              e.dataTransfer.effectAllowed = "move";
            }}
            onDragOver={(e) => {
              e.preventDefault();
              const dragData = dragExerciseRef.current;
              if (!dragData || dragData.stepIdx !== idx) return;
              if (dragData.idx === exIdx) return;
              moveExercise(idx, dragData.idx, exIdx);
              dragExerciseRef.current = { stepIdx: idx, idx: exIdx };
            }}
            onDragEnd={() => {
              dragExerciseRef.current = null;
            }}
          >
            <div className="field">
              <label>Exercise</label>
              <select
                value={ex.exerciseId || (ex.name ? `name:${ex.name}` : "")}
                onChange={async (e) => {
                  const val = e.target.value;
                  if (val === "__add_new") {
                    const newName = await promptUser("Exercise name");
                    if (!newName || !newName.trim()) return;
                    try {
                      const created = await onCreateExercise(newName.trim());
                      updateExercise(idx, exIdx, {
                        name: created.name,
                        exerciseId: created.id,
                      });
                    } catch (err: any) {
                      await notifyUser(
                        err.message || "Unable to create exercise",
                      );
                    }
                    return;
                  }
                  if (!val) {
                    updateExercise(idx, exIdx, {
                      name: "",
                      exerciseId: "",
                    });
                    return;
                  }
                  if (val.startsWith("name:")) return;
                  const selected = catalogById.get(val);
                  if (!selected) return;
                  updateExercise(idx, exIdx, {
                    name: selected.name,
                    exerciseId: selected.id,
                  });
                }}
              >
                <option value="">Select exercise</option>
                {ex.exerciseId && !catalogById.has(ex.exerciseId) && (
                  <option value={ex.exerciseId} disabled>
                    {ex.name || "Unknown exercise"}
                  </option>
                )}
                {!ex.exerciseId && ex.name && (
                  <option value={`name:${ex.name}`} disabled>
                    {ex.name} (unlinked)
                  </option>
                )}
                {catalog.map((c) => (
                  <option
                    key={c.id}
                    value={c.id}
                    className={`exercise-option ${c.isCore ? "core" : "user"}`}
                  >
                    {c.isCore ? `${c.name} ★` : c.name}
                  </option>
                ))}
                <option value="__add_new">+ Add new exercise</option>
              </select>
            </div>
            <div className="field narrow">
              <label>Duration</label>
              <input
                value={ex.amount || ""}
                onChange={(e) =>
                  updateExercise(idx, exIdx, { amount: e.target.value })
                }
                placeholder="e.g. 60s"
              />
            </div>
            <div className="field narrow">
              <label>Transition</label>
              <input
                value={ex.weight || ""}
                onChange={(e) =>
                  updateExercise(idx, exIdx, { weight: e.target.value })
                }
                placeholder="e.g. 10s"
              />
            </div>
            <button
              className="btn icon"
              type="button"
              onClick={() => removeExercise(idx, exIdx)}
              title="Remove exercise"
            >
              ×
            </button>
          </div>
        ))}
        <button
          className="btn outline"
          type="button"
          onClick={() => addExercise(idx)}
        >
          Add Exercise
        </button>
      </div>
    );
  }

  // submit validates and saves the workout.
  const submit = async (e: FormEvent) => {
    e.preventDefault();
    if (!userId) {
      await notifyUser("Select a user first.");
      return;
    }
    const cleanSteps = steps
      .filter((s) => s.name.trim())
      .map((s, idx) => {
        const autoAdvance =
          s.type === "pause" &&
          (s.pauseOptions?.autoAdvance || s.weight === "__auto__");
        const weight =
          s.type === "pause"
            ? autoAdvance
              ? "__auto__"
              : ""
            : (s.weight || "").trim();
        return {
          ...s,
          order: idx,
          weight,
          pauseOptions: autoAdvance ? { autoAdvance: true } : undefined,
          duration: s.duration?.trim() || "",
          exercises:
            s.exercises && s.exercises.length
              ? s.exercises
              : s.type === "set"
                ? [
                    {
                      name: s.name || "Exercise",
                      amount: "",
                      weight: "",
                      exerciseId: s.name
                        ? catalogByName.get(s.name.toLowerCase())?.id
                        : undefined,
                    },
                  ]
                : [],
        };
      });
    if (!cleanSteps.length) {
      await notifyUser("Add at least one step.");
      return;
    }
    try {
      if (editingId && onUpdate) {
        await onUpdate({ id: editingId, name: name.trim(), steps: cleanSteps });
      } else {
        await onSave({ name: name.trim(), steps: cleanSteps });
      }
      setDirty(false);
      onDirtyChange?.(false);
      onToast?.(editingId ? "Workout updated" : "Workout created");
    } catch (err: any) {
      await notifyUser(err?.message || "Unable to save workout");
    }
  };

  return (
    <form className="panel" onSubmit={submit}>
      <div className="panel-header with-close">
        <div>
          <p className="label">{editingId ? "Edit workout" : "New workout"}</p>
        </div>
        <button className="btn primary" type="submit" disabled={!dirty}>
          {editingId ? "Update Workout" : "Save Workout"}
        </button>
        {onClose && (
          <button
            className="btn icon close-btn"
            type="button"
            onClick={onClose}
            title="Close"
          >
            ×
          </button>
        )}
      </div>
      <div className="field">
        <label>Workout name</label>
        <input
          value={name}
          onChange={(e) => {
            setName(e.target.value);
            setDirty(true);
            onDirtyChange?.(true);
          }}
          placeholder="Push Day"
          required
        />
      </div>
      <div className="steps">
        {steps.map((step, idx) => (
          <div
            key={idx}
            className="step-card"
            draggable
            onDragStart={() => (dragIndex.current = idx)}
            onDragOver={(e) => {
              e.preventDefault();
              if (dragIndex.current === null || dragIndex.current === idx)
                return;
              const from = dragIndex.current;
              const to = idx;
              setSteps((prev) => {
                const next = [...prev];
                const [item] = next.splice(from, 1);
                next.splice(to, 0, item);
                setExpandedSteps((exp) => {
                  const updated = new Set<number>();
                  next.forEach((_, i) => {
                    if (exp.has(i === to ? from : i === from ? to : i)) {
                      updated.add(i);
                    }
                  });
                  return updated;
                });
                return next;
              });
              dragIndex.current = idx;
              setDirty(true);
              onDirtyChange?.(true);
            }}
            onDragEnd={() => (dragIndex.current = null)}
          >
            <div
              className="step-top"
              onClick={() =>
                setExpandedSteps((prev) => {
                  const next = new Set(prev);
                  if (next.has(idx)) next.delete(idx);
                  else next.add(idx);
                  return next;
                })
              }
            >
              <span className="badge">{idx + 1}</span>
              <span className="chevron">
                {expandedSteps.has(idx) ? "▾" : "▸"}
              </span>
              <select
                value={step.type}
                onChange={(e) =>
                  updateStep(idx, {
                    type: e.target.value as WorkoutStep["type"],
                  })
                }
              >
                <option
                  value="set"
                  title="Rep-based step with optional target time."
                >
                  Set
                </option>
                <option
                  value="pause"
                  title="Timed break (auto-advance ends it)."
                >
                  Pause
                </option>
                <option
                  value="timed"
                  title="One timed step per exercise plus transitions."
                >
                  Timed Set
                </option>
              </select>
              <button
                className="btn icon"
                type="button"
                onClick={() => moveStep(idx, -1)}
                disabled={idx === 0}
                title="Move up"
              >
                ↑
              </button>
              <button
                className="btn icon"
                type="button"
                onClick={() => moveStep(idx, 1)}
                disabled={idx === steps.length - 1}
                title="Move down"
              >
                ↓
              </button>
              <button
                className="btn icon"
                type="button"
                onClick={() => removeStep(idx)}
                title="Remove set"
              >
                ×
              </button>
            </div>

            <div className="step-preview">
              <div className="muted small">{step.type.toUpperCase()}</div>
              <div className="step-title">{step.name}</div>
              <div className="muted small">
                {step.type !== "timed"
                  ? step.duration ||
                    (step.estimatedSeconds
                      ? `${step.estimatedSeconds}s`
                      : "open")
                  : "Timed block"}
                {step.exercises?.length
                  ? ` • ${step.exercises
                      .map((ex) =>
                        [ex.name, ex.amount, ex.weight]
                          .filter(Boolean)
                          .join(" • "),
                      )
                      .filter(Boolean)
                      .join(" | ")}`
                  : ""}
              </div>
            </div>

            {expandedSteps.has(idx) && (
              <div className="step-details">
                <div className="field">
                  <label>Name</label>
                  <input
                    value={step.name}
                    onChange={(e) => updateStep(idx, { name: e.target.value })}
                    required
                  />
                </div>
                {step.type !== "timed" && (
                  <div className="field">
                    <label>{durationLabel[step.type]}</label>
                    <input
                      value={step.duration || ""}
                      onChange={(e) =>
                        updateStep(idx, { duration: e.target.value })
                      }
                    />
                  </div>
                )}
                {step.type === "pause" && (
                  <label className="field checkbox">
                    <input
                      type="checkbox"
                      checked={
                        step.weight === "__auto__" ||
                        Boolean(step.pauseOptions?.autoAdvance)
                      }
                      onChange={(e) =>
                        updateStep(idx, {
                          pauseOptions: { autoAdvance: e.target.checked },
                          weight: e.target.checked ? "__auto__" : "",
                        })
                      }
                    />
                    <span>Auto-advance when time elapses</span>
                  </label>
                )}
                <div className="field">
                  <label>Sound</label>
                  <select
                    value={step.soundKey || ""}
                    onChange={(e) =>
                      updateStep(idx, { soundKey: e.target.value })
                    }
                  >
                    <option value="">None</option>
                    {sounds.map((s) => (
                      <option key={s.key} value={s.key}>
                        {s.label}
                      </option>
                    ))}
                  </select>
                </div>
                {step.type === "set" && renderStandardExercises(idx, step)}
                {step.type === "timed" && renderTimedExercises(idx, step)}
              </div>
            )}
          </div>
        ))}
      </div>
      <button className="btn outline" type="button" onClick={addStep}>
        Add Step
      </button>
    </form>
  );
}
