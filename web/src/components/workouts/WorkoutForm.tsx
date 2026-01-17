import {
  useEffect,
  useMemo,
  useRef,
  useState,
  FormEvent,
  useCallback,
} from "react";
import type {
  CatalogExercise,
  Exercise,
  SoundOption,
  Workout,
  WorkoutSubset,
  WorkoutStep,
} from "../../types";
import { PauseOptionsField } from "./PauseOptionsField";
import { ArrowDownIcon } from "../icons/ArrowDownIcon";
import { ArrowUpIcon } from "../icons/ArrowUpIcon";
import { TrashIcon } from "../icons/TrashIcon";
import { formatExerciseLine } from "../../utils/format";
import { parseDurationSeconds, isGoDuration } from "../../utils/time";
import {
  EXERCISE_TYPE_REP,
  isDurationExercise,
  normalizeExerciseType,
} from "../../utils/exercise";
import {
  STEP_TYPE_SET,
  isPauseStepType,
  isSetStepType,
  normalizeStepType,
} from "../../utils/step";
import { WorkoutSubsetEditor } from "./WorkoutSubsetEditor";
import { MESSAGES, toErrorMessage } from "../../utils/messages";

const DEFAULT_WORKOUT_NAME = "Push Day";
const KEY_COOLDOWN_MS = 500;

// makeSubsetId creates a stable client id for new subsets.
function makeSubsetId() {
  return `subset-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

// makeStepId creates a stable client id for new steps.
function makeStepId() {
  return `step-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

export type WorkoutFormDefaults = {
  defaultStepSoundKey: string;
  defaultPauseDuration: string;
  defaultPauseSoundKey: string;
  defaultPauseAutoAdvance: boolean;
  repeatRestAfterLastDefault: boolean;
};

export type WorkoutFormServices = {
  onSave: (payload: { name: string; steps: WorkoutStep[] }) => Promise<void>;
  onUpdate?: (payload: {
    id: string;
    name: string;
    steps: WorkoutStep[];
  }) => Promise<void>;
  onCreateExercise: (name: string) => Promise<CatalogExercise>;
  promptUser: (
    message: string,
    defaultValue?: string,
  ) => Promise<string | null>;
  notifyUser: (message: string) => Promise<void>;
  onToast?: (message: string) => void;
};

export type WorkoutFormProps = {
  editingWorkout?: Workout | null;
  sounds: SoundOption[];
  userId: string | null;
  exerciseCatalog?: CatalogExercise[];
  onClose?: () => void;
  defaults: WorkoutFormDefaults;
  services: WorkoutFormServices;
  onDirtyChange?: (dirty: boolean) => void;
};

type DragExercise = { stepIdx: number; subsetIdx: number; idx: number };

export function WorkoutForm({
  editingWorkout,
  sounds,
  userId,
  exerciseCatalog,
  onClose,
  onDirtyChange,
  defaults,
  services,
}: WorkoutFormProps) {
  const {
    onSave,
    onUpdate,
    onCreateExercise,
    promptUser,
    notifyUser,
    onToast,
  } = services;
  const {
    defaultStepSoundKey,
    defaultPauseDuration,
    defaultPauseSoundKey,
    defaultPauseAutoAdvance,
    repeatRestAfterLastDefault,
  } = defaults;
  const [name, setName] = useState(DEFAULT_WORKOUT_NAME);
  const [steps, setSteps] = useState<WorkoutStep[]>([]);
  const [editingId, setEditingId] = useState<string | null>(null);

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

  const dragIndex = useRef<number | null>(null);
  const dragExerciseRef = useRef<DragExercise | null>(null);

  const soundPopoverRef = useRef<HTMLDivElement | null>(null);

  // single “open popover” state to avoid per-row local state explosion
  const [openSoundPicker, setOpenSoundPicker] = useState<{
    stepIdx: number;
    subsetIdx: number;
    exIdx: number;
  } | null>(null);

  const keyCooldownRef = useRef<Record<string, number>>({});
  const buttonCooldownTimersRef = useRef<Record<string, number | null>>({});

  const createBlankSubset = useCallback(
    (): WorkoutSubset => ({
      id: makeSubsetId(),
      name: "",
      duration: "",
      soundKey: defaultStepSoundKey,
      superset: false,
      exercises: [] as Exercise[],
    }),
    [defaultStepSoundKey],
  );

  const markDirty = useCallback(() => {
    setDirty(true);
    onDirtyChange?.(true);
  }, [onDirtyChange]);

  const mutateSubsets = useCallback(
    (
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
    },
    [markDirty],
  );

  // Dismiss the exercise sound popover on outside clicks.
  useEffect(() => {
    if (!openSoundPicker) return;

    const handleClick = (event: MouseEvent) => {
      const target = event.target as HTMLElement | null;
      if (!target) return;
      if (soundPopoverRef.current?.contains(target)) return;
      if (target.closest(".sound-popover-toggle")) return;
      setOpenSoundPicker(null);
    };

    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [openSoundPicker]);

  // Clear button cooldown timers on unmount.
  useEffect(() => {
    return () => {
      Object.values(buttonCooldownTimersRef.current).forEach((timer) => {
        if (timer) window.clearTimeout(timer);
      });
    };
  }, []);

  const now = () => Date.now();

  const tryConsumeKey = useCallback(
    (code: string, button?: HTMLButtonElement | null) => {
      const ts = now();
      const last = keyCooldownRef.current[code];
      if (last && ts - last < KEY_COOLDOWN_MS) return false;

      keyCooldownRef.current[code] = ts;

      if (button) {
        button.classList.add("key-cooldown");
        const existing = buttonCooldownTimersRef.current[code];
        if (existing) window.clearTimeout(existing);

        buttonCooldownTimersRef.current[code] = window.setTimeout(() => {
          button.classList.remove("key-cooldown");
          buttonCooldownTimersRef.current[code] = null;
        }, KEY_COOLDOWN_MS);
      }

      return true;
    },
    [],
  );

  // Normalize incoming workout data when editing or starting fresh.
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

        if (isPauseStepType(step.type)) step.subsets = [];
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
  }, [
    editingWorkout,
    onDirtyChange,
    repeatRestAfterLastDefault,
    createBlankSubset,
  ]);

  // Resolve catalog exercise IDs when the catalog loads.
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

  const toggleRepeatOptions = (idx: number) =>
    setExpandedRepeats((prev) => {
      const next = new Set(prev);
      if (next.has(idx)) next.delete(idx);
      else next.add(idx);
      return next;
    });

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

  const addExercise = (stepIdx: number, subsetIdx: number) => {
    setSteps((prev) =>
      prev.map((step, idx) => {
        if (idx !== stepIdx) return step;

        const subsets = [...(step.subsets || [])];
        const subset = subsets[subsetIdx];
        if (!subset) return step;

        const exercises: Exercise[] = [
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
              updateStep(idx, { repeatCount: Number(e.target.value || 1) })
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
                      updateStep(idx, { repeatRestAfterLast: e.target.checked })
                    }
                  />
                  <span>Pause after last repeat</span>
                </label>
              }
            />
          </>
        )}
      </>
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

        {subsets.map((subset, subsetIdx) => (
          <WorkoutSubsetEditor
            key={subset.id || `${idx}-${subsetIdx}`}
            stepIdx={idx}
            step={step}
            subset={subset}
            subsetIdx={subsetIdx}
            subsetsLength={subsets.length}
            sounds={sounds}
            catalog={catalog}
            addExercise={addExercise}
            removeSubset={removeSubset}
            updateSubset={updateSubset}
            updateExercise={updateExercise}
            removeExercise={removeExercise}
            dragExerciseRef={dragExerciseRef}
            moveExercise={moveExercise}
            soundPopoverRef={soundPopoverRef}
            isSoundOpen={(s, sub, ex) =>
              Boolean(
                openSoundPicker &&
                  openSoundPicker.stepIdx === s &&
                  openSoundPicker.subsetIdx === sub &&
                  openSoundPicker.exIdx === ex,
              )
            }
            setSoundOpen={(s, sub, ex, open) =>
              setOpenSoundPicker(
                open ? { stepIdx: s, subsetIdx: sub, exIdx: ex } : null,
              )
            }
            promptUser={promptUser}
            onCreateExercise={onCreateExercise}
            notifyUser={notifyUser}
          />
        ))}

        {renderRepeatFields(idx, step)}
      </div>
    );
  }

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    if (!userId) {
      await notifyUser("Select a user first.");
      return;
    }

    const cleanSteps: WorkoutStep[] = steps
      .filter((s) => s.name.trim())
      .map((s, idx) => {
        const type = normalizeStepType(s.type);

        const autoAdvance =
          isPauseStepType(type) && s.pauseOptions?.autoAdvance;
        const cleanSubsets = isSetStepType(type)
          ? (s.subsets || []).map((subset) => ({
              id: subset.id || makeSubsetId(),
              name: subset.name?.trim() || "",
              duration: subset.duration?.trim() || "",
              soundKey: subset.soundKey?.trim() || "",
              superset: Boolean(subset.superset),
              exercises: (subset.exercises || [])
                .map((ex) => {
                  const exType = normalizeExerciseType(
                    ex.type,
                  ) as Exercise["type"];
                  const isDur = isDurationExercise(exType);

                  return {
                    exerciseId: ex.exerciseId,
                    name: ex.name.trim(),
                    type: exType,
                    reps:
                      exType === EXERCISE_TYPE_REP ? ex.reps?.trim() || "" : "",
                    weight: ex.weight?.trim() || "",
                    duration: isDur ? ex.duration?.trim() || "" : "",
                    soundKey: ex.soundKey?.trim() || "",
                  };
                })
                .filter((ex) => ex.name.trim()),
            }))
          : [];

        return {
          ...s,
          type,
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
    } catch (err) {
      await notifyUser(toErrorMessage(err, MESSAGES.saveWorkoutFailed));
    }
  };

  // Optional: keyboard QoL (consistent cooldown)
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement | null)?.tagName;
      if (tag === "INPUT" || tag === "SELECT" || tag === "TEXTAREA") return;

      if (e.code === "Enter") {
        if (!dirty) return;
        if (!tryConsumeKey("Enter")) return;
      }
    };

    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [dirty, tryConsumeKey]);

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
                      if (!step.duration)
                        patch.duration = defaultPauseDuration || "";
                      if (!step.soundKey)
                        patch.soundKey =
                          defaultPauseSoundKey || defaultStepSoundKey || "";

                      if (defaultPauseAutoAdvance)
                        patch.pauseOptions = { autoAdvance: true };
                      else patch.pauseOptions = undefined;

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
                  className="btn icon icon-only move"
                  type="button"
                  onClick={() => moveStep(idx, -1)}
                  disabled={idx === 0}
                  title="Move up"
                >
                  <ArrowUpIcon />
                </button>

                <button
                  className="btn icon icon-only move"
                  type="button"
                  onClick={() => moveStep(idx, 1)}
                  disabled={idx === steps.length - 1}
                  title="Move down"
                >
                  <ArrowDownIcon />
                </button>

                <button
                  className="btn icon delete icon-only"
                  type="button"
                  onClick={() => removeStep(idx)}
                  title="Remove step"
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
                        className={
                          (step.duration || "").trim() !== "" &&
                          !isGoDuration((step.duration || "").trim())
                            ? "input-error"
                            : undefined
                        }
                      />
                      {(step.duration || "").trim() !== "" &&
                        !isGoDuration((step.duration || "").trim()) && (
                          <div className="helper error">
                            Use Go duration like 45s or 1m30s
                          </div>
                        )}
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
