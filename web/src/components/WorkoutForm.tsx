import { useEffect, useMemo, useRef, useState, FormEvent } from "react";
import type {
  CatalogExercise,
  Exercise,
  SoundOption,
  Workout,
  WorkoutStep,
} from "../types";
import { ExerciseSelect } from "./ExerciseSelect";
import { PauseOptionsField } from "./PauseOptionsField";
import { parseDurationSeconds } from "../utils/time";

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
  defaultStepSoundKey,
  defaultPauseDuration,
  defaultPauseSoundKey,
  defaultPauseAutoAdvance,
  repeatRestAfterLastDefault,
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
  defaultStepSoundKey: string;
  defaultPauseDuration: string;
  defaultPauseSoundKey: string;
  defaultPauseAutoAdvance: boolean;
  repeatRestAfterLastDefault: boolean;
  onDirtyChange?: (dirty: boolean) => void;
  onToast?: (message: string) => void;
}) {
  const [name, setName] = useState(DEFAULT_WORKOUT_NAME);
  const [steps, setSteps] = useState<WorkoutStep[]>([]);
  const [editingId, setEditingId] = useState<string | null>(null);
  const dragIndex = useRef<number | null>(null);
  const dragExerciseRef = useRef<{ stepIdx: number; idx: number } | null>(null);
  const [expandedSteps, setExpandedSteps] = useState<Set<number>>(new Set([0]));
  const [expandedRepeats, setExpandedRepeats] = useState<Set<number>>(
    new Set(),
  );
  const [repeatRestInputs, setRepeatRestInputs] = useState<string[]>([]);
  const [dirty, setDirty] = useState(false);
  const catalog = exerciseCatalog || [];
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
      setExpandedRepeats(new Set());
      setRepeatRestInputs([]);
      setDirty(false);
      onDirtyChange?.(false);
      return;
    }

    setEditingId(editingWorkout.id);
    setName(editingWorkout.name);
    setSteps(
      (editingWorkout.steps || []).map((s) => ({
        ...s,
        pauseOptions: s.pauseOptions,
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
        repeatCount: s.repeatCount || 1,
        repeatRestSeconds: s.repeatRestSeconds || 0,
        repeatRestAfterLast:
          typeof s.repeatRestAfterLast === "boolean"
            ? s.repeatRestAfterLast
            : repeatRestAfterLastDefault,
        repeatRestSoundKey: s.repeatRestSoundKey || "",
        repeatRestAutoAdvance:
          typeof s.repeatRestAutoAdvance === "boolean"
            ? s.repeatRestAutoAdvance
            : true,
      })),
    );
    setRepeatRestInputs(
      (editingWorkout.steps || []).map((s) =>
        s.repeatRestSeconds ? `${s.repeatRestSeconds}s` : "",
      ),
    );
    setDirty(false);
    setExpandedRepeats(new Set());
    onDirtyChange?.(false);
  }, [editingWorkout, onDirtyChange, repeatRestAfterLastDefault]);

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
        soundKey: defaultStepSoundKey,
        repeatCount: 1,
        repeatRestSeconds: 0,
        repeatRestAfterLast: repeatRestAfterLastDefault,
        repeatRestSoundKey: "",
        repeatRestAutoAdvance: true,
      };
      const next = [...prev, newStep];
      setExpandedSteps((exp) => new Set(exp).add(next.length - 1));
      setRepeatRestInputs((inputs) => [...inputs, ""]);
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
    setExpandedRepeats((prev) => {
      const next = new Set(prev);
      next.delete(idx);
      return next;
    });
    setRepeatRestInputs((prev) => prev.filter((_, i) => i !== idx));
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
    setRepeatRestInputs((prev) => {
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

  // toggleRepeatOptions expands or collapses the repeat settings for a step.
  const toggleRepeatOptions = (idx: number) =>
    setExpandedRepeats((prev) => {
      const next = new Set(prev);
      if (next.has(idx)) {
        next.delete(idx);
      } else {
        next.add(idx);
      }
      return next;
    });

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

  // renderRepeatToggle shows the repeat options trigger next to action buttons.
  function renderRepeatToggle(idx: number, step: WorkoutStep) {
    return (
      <>
        <button
          className="btn subtle"
          type="button"
          onClick={() => toggleRepeatOptions(idx)}
        >
          {expandedRepeats.has(idx) ? "Hide repeat options" : "Repeat options"}
        </button>
        <span className="muted small">
          {step.repeatCount && step.repeatCount > 1
            ? `Repeats ${step.repeatCount}x`
            : "No repeats"}
        </span>
      </>
    );
  }

  // renderRepeatFields shows expanded repeat settings under the action row.
  function renderRepeatFields(idx: number, step: WorkoutStep) {
    if (!expandedRepeats.has(idx)) return null;
    return (
      <>
        <div className="field">
          <label>Repeat count</label>
          <input
            type="number"
            min={1}
            value={step.repeatCount ?? 1}
            onChange={(e) =>
              updateStep(idx, {
                repeatCount: Number(e.target.value || 1),
              })
            }
          />
        </div>
        {Boolean(step.repeatCount && step.repeatCount > 1) && (
          <>
            <div className="divider" />
            <div className="label">Repeat pause</div>
            <div className="muted small hint">
              Configure the break between repeat rounds.
            </div>
            <div className="field">
              <label>{durationLabel.pause}</label>
              <input
                value={
                  repeatRestInputs[idx] ??
                  (step.repeatRestSeconds ? `${step.repeatRestSeconds}s` : "")
                }
                onChange={(e) => {
                  const value = e.target.value;
                  setRepeatRestInputs((prev) => {
                    const next = [...prev];
                    next[idx] = value;
                    return next;
                  });
                  updateStep(idx, {
                    repeatRestSeconds: parseDurationSeconds(value),
                  });
                }}
              />
            </div>
          </>
        )}
        {Boolean(step.repeatCount && step.repeatCount > 1) && (
          <PauseOptionsField
            autoAdvance={Boolean(step.repeatRestAutoAdvance)}
            soundKey={step.repeatRestSoundKey || ""}
            sounds={sounds}
            onAutoAdvanceChange={(value) =>
              updateStep(idx, { repeatRestAutoAdvance: value })
            }
            onSoundChange={(value) =>
              updateStep(idx, { repeatRestSoundKey: value })
            }
            extra={
              <label className="field checkbox">
                <input
                  type="checkbox"
                  checked={Boolean(step.repeatRestAfterLast)}
                  onChange={(e) =>
                    updateStep(idx, {
                      repeatRestAfterLast: e.target.checked,
                    })
                  }
                />
                <span>Pause after last repeat</span>
              </label>
            }
          />
        )}
      </>
    );
  }

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
              <ExerciseSelect
                catalog={catalog}
                value={{ exerciseId: ex.exerciseId, name: ex.name }}
                onSelect={(selected) =>
                  updateExercise(idx, exIdx, {
                    name: selected.name,
                    exerciseId: selected.id,
                  })
                }
                onClear={() =>
                  updateExercise(idx, exIdx, { name: "", exerciseId: "" })
                }
                onAddNew={async () => {
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
                }}
              />
            </div>
            <div className="field">
              <label>Amount / Reps</label>
              <input
                value={ex.amount || ""}
                onChange={(e) =>
                  updateExercise(idx, exIdx, { amount: e.target.value })
                }
                placeholder="12 reps"
              />
            </div>
            <div className="field">
              <label>Weight</label>
              <input
                value={ex.weight || ""}
                onChange={(e) =>
                  updateExercise(idx, exIdx, { weight: e.target.value })
                }
                placeholder="50kg"
              />
            </div>
            <div className="field action">
              <label>Action</label>
              <button
                className="btn icon"
                type="button"
                onClick={() => removeExercise(idx, exIdx)}
                title="Remove exercise"
              >
                ×
              </button>
            </div>
          </div>
        ))}
        <div className="btn-group">
          <button
            className="btn outline"
            type="button"
            onClick={() => addExercise(idx)}
          >
            Add Exercise
          </button>
          {renderRepeatToggle(idx, step)}
        </div>
        {renderRepeatFields(idx, step)}
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
              <ExerciseSelect
                catalog={catalog}
                value={{ exerciseId: ex.exerciseId, name: ex.name }}
                onSelect={(selected) =>
                  updateExercise(idx, exIdx, {
                    name: selected.name,
                    exerciseId: selected.id,
                  })
                }
                onClear={() =>
                  updateExercise(idx, exIdx, { name: "", exerciseId: "" })
                }
                onAddNew={async () => {
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
                }}
              />
            </div>
            <div className="field">
              <label>Duration</label>
              <input
                value={ex.amount || ""}
                onChange={(e) =>
                  updateExercise(idx, exIdx, { amount: e.target.value })
                }
                placeholder="e.g. 60s"
              />
            </div>
            <div className="field">
              <label>Transition</label>
              <input
                value={ex.weight || ""}
                onChange={(e) =>
                  updateExercise(idx, exIdx, { weight: e.target.value })
                }
                placeholder="e.g. 10s"
              />
            </div>
            <div className="field action">
              <button
                className="btn icon"
                type="button"
                onClick={() => removeExercise(idx, exIdx)}
                title="Remove exercise"
              >
                ×
              </button>
            </div>
          </div>
        ))}
        <div className="btn-group">
          <button
            className="btn outline"
            type="button"
            onClick={() => addExercise(idx)}
          >
            Add Exercise
          </button>
          {renderRepeatToggle(idx, step)}
        </div>
        {renderRepeatFields(idx, step)}
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
        const autoAdvance = s.type === "pause" && s.pauseOptions?.autoAdvance;
        return {
          ...s,
          order: idx,
          pauseOptions: autoAdvance ? { autoAdvance: true } : undefined,
          duration: s.duration?.trim() || "",
          repeatCount: Math.max(1, Math.floor(Number(s.repeatCount) || 1)),
          repeatRestSeconds: Math.max(
            0,
            Math.floor(Number(s.repeatRestSeconds) || 0),
          ),
          repeatRestAfterLast: Boolean(s.repeatRestAfterLast),
          repeatRestSoundKey:
            s.repeatRestSeconds && (s.repeatRestSoundKey || "").trim()
              ? (s.repeatRestSoundKey || "").trim()
              : "",
          repeatRestAutoAdvance:
            Boolean(s.repeatRestSeconds) && Boolean(s.repeatRestAutoAdvance),
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
      {/* Workout steps list */}
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
              setRepeatRestInputs((prev) => {
                const next = [...prev];
                const [item] = next.splice(from, 1);
                next.splice(to, 0, item);
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
                {expandedSteps.has(idx) ? "▼" : "▶"}
              </span>
              <select
                value={step.type}
                onChange={(e) => {
                  const nextType = e.target.value as WorkoutStep["type"];
                  const patch: Partial<WorkoutStep> = { type: nextType };
                  if (nextType === "pause") {
                    if (!step.duration) {
                      patch.duration = defaultPauseDuration || "";
                    }
                    if (!step.soundKey) {
                      patch.soundKey =
                        defaultPauseSoundKey || defaultStepSoundKey || "";
                    }
                    if (defaultPauseAutoAdvance) {
                      patch.pauseOptions = { autoAdvance: true };
                    } else {
                      patch.pauseOptions = undefined;
                    }
                  } else {
                    patch.pauseOptions = undefined;
                  }
                  updateStep(idx, patch);
                }}
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
              <div className="muted small hint">{step.type.toUpperCase()}</div>
              <div className="step-title">{step.name}</div>
              <div className="muted small">
                {step.type !== "timed"
                  ? step.duration ||
                    (step.estimatedSeconds
                      ? `${step.estimatedSeconds}s`
                      : "open")
                  : "Timed block"}
                {step.repeatCount && step.repeatCount > 1
                  ? ` • repeats ${step.repeatCount}x`
                  : ""}
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
                <div className="field spaced">
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
                {step.type === "timed" && (
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
                )}
                {step.type === "pause" && (
                  <PauseOptionsField
                    autoAdvance={Boolean(step.pauseOptions?.autoAdvance)}
                    soundKey={step.soundKey || ""}
                    sounds={sounds}
                    onAutoAdvanceChange={(value) =>
                      updateStep(idx, {
                        pauseOptions: { autoAdvance: value },
                      })
                    }
                    onSoundChange={(value) =>
                      updateStep(idx, { soundKey: value })
                    }
                  />
                )}
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
