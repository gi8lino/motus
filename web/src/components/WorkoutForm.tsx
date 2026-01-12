import { useEffect, useMemo, useRef, useState, FormEvent } from "react";
import type {
  CatalogExercise,
  Exercise,
  SoundOption,
  Workout,
  WorkoutSubset,
  WorkoutStep,
} from "../types";
import { ExerciseSelect } from "./ExerciseSelect";
import { PauseOptionsField } from "./PauseOptionsField";
import { SoundIcon } from "./icons/SoundIcon";
import { TrashIcon } from "./icons/TrashIcon";
import { formatExerciseLine } from "../utils/format";
import { parseDurationSeconds, isGoDuration } from "../utils/time";
import { isRepRange } from "../utils/validation";
import {
  EXERCISE_TYPE_COUNTDOWN,
  EXERCISE_TYPE_REP,
  EXERCISE_TYPE_STOPWATCH,
  isDurationExercise,
  normalizeExerciseType,
} from "../utils/exercise";
import {
  STEP_TYPE_SET,
  isPauseStepType,
  isSetStepType,
  normalizeStepType,
} from "../utils/step";

const DEFAULT_WORKOUT_NAME = "Push Day";

// makeSubsetId creates a stable client id for new subsets.
function makeSubsetId() {
  return `subset-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

// makeStepId creates a stable client id for new steps.
function makeStepId() {
  return `step-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

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
  const dragExerciseRef = useRef<{
    stepIdx: number;
    subsetIdx: number;
    idx: number;
  } | null>(null);
  const [expandedSteps, setExpandedSteps] = useState<Set<number>>(new Set([0]));
  const [expandedRepeats, setExpandedRepeats] = useState<Set<number>>(
    new Set(),
  );
  const [repeatRestInputs, setRepeatRestInputs] = useState<string[]>([]);
  const [dirty, setDirty] = useState(false);
  const [exerciseSoundPicker, setExerciseSoundPicker] = useState<{
    stepIdx: number;
    subsetIdx: number;
    exIdx: number;
  } | null>(null);
  const soundPopoverRef = useRef<HTMLDivElement | null>(null);
  const catalog = exerciseCatalog || [];
  const catalogByName = useMemo(
    () => new Map(catalog.map((entry) => [entry.name.toLowerCase(), entry])),
    [catalog],
  );

  const createBlankSubset = (): WorkoutSubset => ({
    id: makeSubsetId(),
    name: "",
    duration: "",
    soundKey: defaultStepSoundKey,
    superset: false,
    exercises: [] as Exercise[],
  });

  const markDirty = () => {
    setDirty(true);
    onDirtyChange?.(true);
  };

  const mutateSubsets = (
    stepIdx: number,
    mutator: (subsets: WorkoutSubset[]) => WorkoutSubset[],
  ) => {
    setSteps((prev) =>
      prev.map((step, idx) => {
        if (idx !== stepIdx) return step;
        const subsets = mutator(step.subsets || []);
        return { ...step, subsets };
      }),
    );
    markDirty();
  };

  useEffect(() => {
    // Close the exercise sound popover when clicking outside.
    if (!exerciseSoundPicker) return;
    const handleClick = (event: MouseEvent) => {
      const target = event.target as HTMLElement | null;
      if (!target) return;
      if (soundPopoverRef.current?.contains(target)) return;
      if (target.closest(".sound-popover-toggle")) return;
      setExerciseSoundPicker(null);
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [exerciseSoundPicker]);

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
      (editingWorkout.steps || []).map((s) => {
        const subsets =
          s.subsets && s.subsets.length
            ? s.subsets
            : isSetStepType(s.type)
              ? [createBlankSubset()]
              : [];
        const normalizedSubsets = subsets.map((subset) => ({
          ...subset,
          id: subset.id || makeSubsetId(),
          name: subset.name || "",
          duration:
            subset.duration ||
            (subset.estimatedSeconds ? `${subset.estimatedSeconds}s` : ""),
          soundKey: subset.soundKey || "",
          superset: Boolean(subset.superset),
          exercises: (subset.exercises || []).map((ex) => ({
            exerciseId: ex.exerciseId,
            name: ex.name,
            type: (ex.type || "rep") as Exercise["type"],
            reps: ex.reps || "",
            weight: ex.weight || "",
            duration: ex.duration || "",
            soundKey: ex.soundKey || "",
          })),
        }));

        const stepType = normalizeStepType(s.type);
        const step: WorkoutStep = {
          ...s,
          id: s.id || makeStepId(),
          pauseOptions: s.pauseOptions,
          type: stepType,
          duration: isPauseStepType(stepType)
            ? s.duration || (s.estimatedSeconds ? `${s.estimatedSeconds}s` : "")
            : "",
          soundKey: s.soundKey || "",
          subsets: normalizedSubsets,
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
        };

        if (isPauseStepType(step.type)) {
          step.subsets = [];
        }
        return step;
      }),
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
        subsets: (step.subsets || []).map((subset) => ({
          ...subset,
          exercises: (subset.exercises || []).map((ex) => {
            if (ex.exerciseId || !ex.name) return ex;
            const resolved = catalogByName.get(ex.name.toLowerCase());
            if (!resolved) return ex;
            return { ...ex, exerciseId: resolved.id };
          }),
        })),
      })),
    );
  }, [catalogByName, catalog.length]);

  // addStep appends a new default step and expands it for editing.
  const addStep = () =>
    setSteps((prev) => {
      const newStep: WorkoutStep = {
        id: makeStepId(),
        type: STEP_TYPE_SET,
        name: `Step ${prev.length + 1}`,
        repeatCount: 1,
        repeatRestSeconds: 0,
        repeatRestAfterLast: repeatRestAfterLastDefault,
        repeatRestSoundKey: "",
        repeatRestAutoAdvance: true,
        subsets: [createBlankSubset()],
      };
      const next = [...prev, newStep];
      setExpandedSteps((exp) => new Set(exp).add(next.length - 1));
      setRepeatRestInputs((inputs) => [...inputs, ""]);
      markDirty();
      return next;
    });

  // updateStep merges a partial update into a step at an index.
  const updateStep = (idx: number, patch: Partial<WorkoutStep>) => {
    setSteps((prev) =>
      prev.map((step, i) => (i === idx ? { ...step, ...patch } : step)),
    );
    markDirty();
  };

  const updateSubset = (
    stepIdx: number,
    subsetIdx: number,
    patch: Partial<WorkoutSubset>,
  ) => {
    mutateSubsets(stepIdx, (subsets) => {
      const clone = [...subsets];
      if (!clone[subsetIdx]) return subsets;
      clone[subsetIdx] = { ...clone[subsetIdx], ...patch };
      return clone;
    });
  };

  const addSubset = (stepIdx: number) => {
    mutateSubsets(stepIdx, (subsets) => [...subsets, createBlankSubset()]);
  };

  const removeSubset = (stepIdx: number, subsetIdx: number) => {
    mutateSubsets(stepIdx, (subsets) => {
      if (subsets.length <= 1) return subsets;
      const next = [...subsets];
      next.splice(subsetIdx, 1);
      return next;
    });
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
    markDirty();
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
    markDirty();
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
  const moveExercise = (
    stepIdx: number,
    subsetIdx: number,
    from: number,
    to: number,
  ) => {
    setSteps((prev) =>
      prev.map((step, idx) => {
        if (idx !== stepIdx) return step;
        const subsets = [...(step.subsets || [])];
        const subset = subsets[subsetIdx];
        if (!subset) return step;
        const updated = [...(subset.exercises || [])];
        const [item] = updated.splice(from, 1);
        updated.splice(to, 0, item);
        subsets[subsetIdx] = { ...subset, exercises: updated };
        return { ...step, subsets };
      }),
    );
    markDirty();
  };

  // updateExercise mutates an exercise at an index within a step.
  const updateExercise = (
    stepIdx: number,
    subsetIdx: number,
    exIdx: number,
    patch: Partial<Exercise>,
  ) => {
    setSteps((prev) =>
      prev.map((step, idx) => {
        if (idx !== stepIdx) return step;
        const subsets = [...(step.subsets || [])];
        const subset = subsets[subsetIdx];
        if (!subset) return step;
        const updated = [...(subset.exercises || [])];
        updated[exIdx] = { ...updated[exIdx], ...patch };
        subsets[subsetIdx] = { ...subset, exercises: updated };
        return { ...step, subsets };
      }),
    );
    markDirty();
  };

  // addExercise appends a blank exercise row to a step.
  const addExercise = (stepIdx: number, subsetIdx: number) => {
    setSteps((prev) =>
      prev.map((step, idx) => {
        if (idx !== stepIdx) return step;
        const subsets = [...(step.subsets || [])];
        const subset = subsets[subsetIdx];
        if (!subset) return step;
        const exercises = [
          ...(subset.exercises || []),
          {
            name: "",
            reps: "",
            weight: "",
            duration: "",
            exerciseId: "",
            type: "rep" as Exercise["type"],
            soundKey: "",
          },
        ];
        subsets[subsetIdx] = { ...subset, exercises };
        return { ...step, subsets };
      }),
    );
    markDirty();
  };

  // removeExercise deletes an exercise row.
  const removeExercise = (
    stepIdx: number,
    subsetIdx: number,
    exIdx: number,
  ) => {
    setSteps((prev) =>
      prev.map((step, idx) => {
        if (idx !== stepIdx) return step;
        const subsets = [...(step.subsets || [])];
        const subset = subsets[subsetIdx];
        if (!subset) return step;
        const updated = [...(subset.exercises || [])];
        updated.splice(exIdx, 1);
        subsets[subsetIdx] = { ...subset, exercises: updated };
        return { ...step, subsets };
      }),
    );
    markDirty();
  };

  const durationLabel = useMemo(
    () => ({
      pause: "Duration (Go style, e.g. 45s, 1m30s)",
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

  function renderExerciseRow(
    stepIdx: number,
    subsetIdx: number,
    exIdx: number,
    ex: Exercise,
  ) {
    const kind = normalizeExerciseType(ex.type);
    const showDuration = isDurationExercise(kind);
    const amountLabel = showDuration ? "Duration" : "Reps";
    const amountPlaceholder = showDuration ? "e.g. 45s" : "12";
    const repsValue = (ex.reps || "").trim();
    const durationValue = (ex.duration || "").trim();
    const soundKey = (ex.soundKey || "").trim();
    const soundLabel =
      sounds.find((sound) => sound.key === soundKey)?.label || "Sound";
    const soundSummary = soundKey ? soundLabel : "Subset sound";
    const soundOpen =
      exerciseSoundPicker?.stepIdx === stepIdx &&
      exerciseSoundPicker?.subsetIdx === subsetIdx &&
      exerciseSoundPicker?.exIdx === exIdx;
    const repsInvalid =
      !showDuration && repsValue !== "" && !isRepRange(repsValue);
    const durationInvalid =
      showDuration && durationValue !== "" && !isGoDuration(durationValue);

    return (
      <div
        key={exIdx}
        className="exercise-row"
        draggable
        onDragStart={(e) => {
          e.stopPropagation();
          dragExerciseRef.current = { stepIdx, subsetIdx, idx: exIdx };
          e.dataTransfer.effectAllowed = "move";
        }}
        onDragOver={(e) => {
          e.preventDefault();
          const dragData = dragExerciseRef.current;
          if (
            !dragData ||
            dragData.stepIdx !== stepIdx ||
            dragData.subsetIdx !== subsetIdx
          ) {
            return;
          }
          if (dragData.idx === exIdx) return;
          moveExercise(stepIdx, subsetIdx, dragData.idx, exIdx);
          dragExerciseRef.current = { stepIdx, subsetIdx, idx: exIdx };
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
              updateExercise(stepIdx, subsetIdx, exIdx, {
                name: selected.name,
                exerciseId: selected.id,
              })
            }
            onClear={() =>
              updateExercise(stepIdx, subsetIdx, exIdx, {
                name: "",
                exerciseId: "",
              })
            }
            onAddNew={async () => {
              const newName = await promptUser("Exercise name");
              if (!newName || !newName.trim()) return;
              try {
                const created = await onCreateExercise(newName.trim());
                updateExercise(stepIdx, subsetIdx, exIdx, {
                  name: created.name,
                  exerciseId: created.id,
                });
              } catch (err: any) {
                await notifyUser(err.message || "Unable to create exercise");
              }
            }}
          />
        </div>
        <div className="field compact">
          <label>Exercise type</label>
          <select
            value={kind}
            onChange={(e) =>
              updateExercise(stepIdx, subsetIdx, exIdx, {
                type: e.target.value as Exercise["type"],
              })
            }
          >
            <option value={EXERCISE_TYPE_REP}>Reps</option>
            <option value={EXERCISE_TYPE_STOPWATCH}>Stopwatch</option>
            <option value={EXERCISE_TYPE_COUNTDOWN}>Countdown</option>
          </select>
        </div>
        <div className="field compact">
          <label>{amountLabel}</label>
          <input
            value={showDuration ? ex.duration || "" : ex.reps || ""}
            onChange={(e) =>
              updateExercise(stepIdx, subsetIdx, exIdx, {
                ...(showDuration
                  ? { duration: e.target.value }
                  : { reps: e.target.value }),
              })
            }
            className={
              repsInvalid || durationInvalid ? "input-error" : undefined
            }
            placeholder={amountPlaceholder}
          />
          {repsInvalid && <div className="helper error">Use 8 or 8-10</div>}
          {durationInvalid && (
            <div className="helper error">
              Use Go duration like 45s or 1m30s
            </div>
          )}
        </div>
        <div className="field compact">
          <label>Weight</label>
          <input
            value={ex.weight || ""}
            onChange={(e) =>
              updateExercise(stepIdx, subsetIdx, exIdx, {
                weight: e.target.value,
              })
            }
            placeholder="10kg"
          />
        </div>
        <div className="field action compact sound">
          <label>Sound</label>
          <button
            className={[
              "btn",
              "subtle",
              "tiny",
              "sound-popover-toggle",
              "exercise-sound-button",
              soundKey ? "is-override" : "",
            ]
              .filter(Boolean)
              .join(" ")}
            type="button"
            title={`Sound: ${soundSummary}`}
            data-label={soundSummary}
            onClick={() =>
              setExerciseSoundPicker((prev) =>
                prev &&
                prev.stepIdx === stepIdx &&
                prev.subsetIdx === subsetIdx &&
                prev.exIdx === exIdx
                  ? null
                  : { stepIdx, subsetIdx, exIdx },
              )
            }
          >
            <SoundIcon />
          </button>
          {soundOpen && (
            <div ref={soundPopoverRef} className="sound-popover">
              <div className="sound-popover-title">Exercise sound</div>
              <button
                className="sound-popover-option"
                type="button"
                onClick={() => {
                  updateExercise(stepIdx, subsetIdx, exIdx, { soundKey: "" });
                  setExerciseSoundPicker(null);
                }}
              >
                Use subset sound
              </button>
              {sounds.map((sound) => (
                <button
                  key={sound.key}
                  className="sound-popover-option"
                  type="button"
                  onClick={() => {
                    updateExercise(stepIdx, subsetIdx, exIdx, {
                      soundKey: sound.key,
                    });
                    setExerciseSoundPicker(null);
                  }}
                >
                  {sound.label}
                </button>
              ))}
            </div>
          )}
        </div>
        <div className="field action compact">
          <button
            className="btn icon delete mobile-full"
            type="button"
            onClick={() => removeExercise(stepIdx, subsetIdx, exIdx)}
            title="Remove exercise"
          >
            <span className="desktop-only">
              <TrashIcon />
            </span>
            <span className="mobile-only">Remove exercise</span>
          </button>
        </div>
      </div>
    );
  }

  function renderSubsetEditor(idx: number, step: WorkoutStep) {
    const subsets = step.subsets || [];
    return (
      <div className="stack">
        <div className="field spaced">
          <label>Subsets</label>
          <div className="btn-group">
            <button
              className="btn outline"
              type="button"
              onClick={() => addSubset(idx)}
            >
              Add subset
            </button>
            {renderRepeatToggle(idx, step)}
          </div>
        </div>

        {!subsets.length && (
          <div className="muted small">
            Add a subset to define the exercises and timing for this set.
          </div>
        )}

        {subsets.map((subset, subsetIdx) => {
          const targetDurationValue = (subset.duration || "").trim();
          const targetDurationInvalid =
            targetDurationValue !== "" && !isGoDuration(targetDurationValue);
          const hasMultiple = subsets.length > 1;
          const subsetLabel = subset.name?.trim() || `Subset ${subsetIdx + 1}`;
          const subsetExercises = subset.exercises || [];
          return (
            <div
              key={subset.id || `${idx}-${subsetIdx}`}
              className="stack"
              style={{
                gap: 8,
                borderBottom: "1px solid #eee",
                paddingBottom: 16,
              }}
            >
              {hasMultiple && (
                <div className="set-header">
                  <strong>{subsetLabel}</strong>
                  <div className="subset-actions">
                    {subset.superset ? (
                      <span className="set-badge superset">Superset</span>
                    ) : null}
                    <button
                      className="btn icon delete icon-only"
                      type="button"
                      onClick={() => removeSubset(idx, subsetIdx)}
                      disabled={subsets.length <= 1}
                      title="Remove subset"
                    >
                      <TrashIcon />
                    </button>
                  </div>
                </div>
              )}
              <div className="field">
                <label>Name (optional)</label>
                <input
                  value={subset.name}
                  onChange={(event) =>
                    updateSubset(idx, subsetIdx, { name: event.target.value })
                  }
                />
              </div>
              <div className="field">
                <label>Target time (optional, e.g. 45s)</label>
                <input
                  value={subset.duration || ""}
                  onChange={(event) =>
                    updateSubset(idx, subsetIdx, {
                      duration: event.target.value,
                    })
                  }
                  className={targetDurationInvalid ? "input-error" : undefined}
                />
                {targetDurationInvalid && (
                  <div className="helper error">
                    Use Go duration like 45s or 1m30s
                  </div>
                )}
              </div>
              <div className="field">
                <label>Sound</label>
                <select
                  value={subset.soundKey || ""}
                  onChange={(event) =>
                    updateSubset(idx, subsetIdx, {
                      soundKey: event.target.value,
                    })
                  }
                >
                  <option value="">None</option>
                  {sounds.map((sound) => (
                    <option key={sound.key} value={sound.key}>
                      {sound.label}
                    </option>
                  ))}
                </select>
              </div>
              <label className="field checkbox">
                <input
                  type="checkbox"
                  checked={Boolean(subset.superset)}
                  onChange={(event) =>
                    updateSubset(idx, subsetIdx, {
                      superset: event.target.checked,
                    })
                  }
                />
                <span>
                  Treat this subset as a superset block (Next moves to the next
                  subset).
                </span>
              </label>

              {subsetExercises.length ? (
                subsetExercises.map((exercise, exIdx) =>
                  renderExerciseRow(idx, subsetIdx, exIdx, exercise),
                )
              ) : (
                <div className="muted small">
                  Add at least one exercise for this subset.
                </div>
              )}

              <div className="btn-group">
                <button
                  className="btn outline"
                  type="button"
                  onClick={() => addExercise(idx, subsetIdx)}
                >
                  Add Exercise
                </button>
              </div>
            </div>
          );
        })}

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

    const cleanSteps: WorkoutStep[] = steps
      .filter((s) => s.name.trim())
      .map((s, idx) => {
        const autoAdvance =
          isPauseStepType(s.type) && s.pauseOptions?.autoAdvance;
        const cleanSubsets = isSetStepType(s.type)
          ? (s.subsets || []).map((subset) => ({
              id: subset.id || makeSubsetId(),
              name: subset.name?.trim() || "",
              duration: subset.duration?.trim() || "",
              soundKey: subset.soundKey?.trim() || "",
              superset: Boolean(subset.superset),
              exercises: (subset.exercises || []).map((ex) => {
                const type = normalizeExerciseType(ex.type) as Exercise["type"];
                return {
                  exerciseId: ex.exerciseId,
                  name: ex.name.trim(),
                  type,
                  reps: type === EXERCISE_TYPE_REP ? ex.reps?.trim() || "" : "",
                  weight: ex.weight?.trim() || "",
                  duration:
                    type === EXERCISE_TYPE_REP ? "" : ex.duration?.trim() || "",
                  soundKey: ex.soundKey?.trim() || "",
                };
              }),
            }))
          : [];
        return {
          ...s,
          order: idx,
          pauseOptions: autoAdvance ? { autoAdvance: true } : undefined,
          duration: s.duration?.trim() || "",
          soundKey: s.soundKey?.trim() || "",
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
          subsets: cleanSubsets,
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
            markDirty();
          }}
          placeholder="Push Day"
          required
        />
      </div>
      {/* Workout steps list */}
      <div className="steps">
        {steps.map((step, idx) => {
          const exerciseSummary = (step.subsets || [])
            .flatMap((subset) => subset.exercises || [])
            .map((ex) => formatExerciseLine(ex))
            .filter(Boolean)
            .join(" | ");
          const subsetNames = (step.subsets || [])
            .map((subset) => subset.name?.trim())
            .filter(Boolean)
            .join(" | ");
          return (
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
                markDirty();
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
                      patch.subsets = [];
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
                  className="btn icon delete"
                  type="button"
                  onClick={() => removeStep(idx)}
                  title="Remove set"
                >
                  <TrashIcon />
                </button>
              </div>

              <div className="step-preview">
                <div className="step-title">{step.name}</div>
                <div className="muted small">
                  {step.duration ||
                    (step.estimatedSeconds
                      ? `${step.estimatedSeconds}s`
                      : "open")}
                  {step.repeatCount && step.repeatCount > 1
                    ? ` • repeats ${step.repeatCount}x`
                    : ""}
                  {subsetNames ? ` • ${subsetNames}` : ""}
                  {exerciseSummary ? ` • ${exerciseSummary}` : ""}
                </div>
              </div>

              {expandedSteps.has(idx) && (
                <div className="step-details">
                  <div className="field spaced">
                    <label>
                      {isPauseStepType(step.type) ? "Name (optional)" : "Name"}
                    </label>
                    <input
                      value={step.name}
                      onChange={(e) =>
                        updateStep(idx, { name: e.target.value })
                      }
                      placeholder={isPauseStepType(step.type) ? "Pause" : ""}
                      required={!isPauseStepType(step.type)}
                    />
                  </div>
                  {isPauseStepType(step.type) && (
                    <div className="field">
                      <label>{durationLabel.pause}</label>
                      <input
                        value={step.duration || ""}
                        onChange={(e) =>
                          updateStep(idx, { duration: e.target.value })
                        }
                      />
                    </div>
                  )}
                  {isPauseStepType(step.type) && (
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
                  {isSetStepType(step.type) && renderSubsetEditor(idx, step)}
                </div>
              )}
            </div>
          );
        })}
      </div>
      <button className="btn outline" type="button" onClick={addStep}>
        Add Step
      </button>
    </form>
  );
}
