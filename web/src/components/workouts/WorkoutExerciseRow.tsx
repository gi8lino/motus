import type { CatalogExercise, Exercise, SoundOption } from "../../types";
import { ExerciseSelect } from "./ExerciseSelect";
import { SoundIcon } from "../icons/SoundIcon";
import { TrashIcon } from "../icons/TrashIcon";
import { isGoDuration } from "../../utils/time";
import { isRepRange } from "../../utils/validation";
import {
  EXERCISE_TYPE_COUNTDOWN,
  EXERCISE_TYPE_REP,
  EXERCISE_TYPE_STOPWATCH,
  isDurationExercise,
  normalizeExerciseType,
} from "../../utils/exercise";

type MutableRef<T> = { current: T };
type DragExercise = { stepIdx: number; subsetIdx: number; idx: number };

type WorkoutExerciseRowProps = {
  stepIdx: number;
  subsetIdx: number;
  exIdx: number;
  ex: Exercise;

  catalog: CatalogExercise[];
  sounds: SoundOption[];

  trainingKey: string;
  soundOpen: boolean;
  setSoundOpen: (open: boolean) => void;

  soundPopoverRef: React.RefObject<HTMLDivElement | null>;

  dragExerciseRef: MutableRef<DragExercise | null>;
  moveExercise: (
    stepIdx: number,
    subsetIdx: number,
    from: number,
    to: number,
  ) => void;

  updateExercise: (
    stepIdx: number,
    subsetIdx: number,
    exIdx: number,
    patch: Partial<Exercise>,
  ) => void;
  removeExercise: (stepIdx: number, subsetIdx: number, exIdx: number) => void;

  promptUser: (
    message: string,
    defaultValue?: string,
  ) => Promise<string | null>;
  onCreateExercise: (name: string) => Promise<CatalogExercise>;
  notifyUser: (message: string) => Promise<void>;
};

export function WorkoutExerciseRow({
  stepIdx,
  subsetIdx,
  exIdx,
  ex,
  catalog,
  sounds,
  trainingKey,
  soundOpen,
  setSoundOpen,
  soundPopoverRef,
  dragExerciseRef,
  moveExercise,
  updateExercise,
  removeExercise,
  promptUser,
  onCreateExercise,
  notifyUser,
}: WorkoutExerciseRowProps) {
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

  const repsInvalid =
    !showDuration && repsValue !== "" && !isRepRange(repsValue);
  const durationInvalid =
    showDuration && durationValue !== "" && !isGoDuration(durationValue);

  return (
    <div
      key={`${trainingKey}-${exIdx}`}
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
        if (!dragData) return;
        if (dragData.stepIdx !== stepIdx || dragData.subsetIdx !== subsetIdx)
          return;
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
              await notifyUser(err?.message || "Unable to create exercise");
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
          className={repsInvalid || durationInvalid ? "input-error" : undefined}
          placeholder={amountPlaceholder}
        />
        {repsInvalid && <div className="helper error">Use 8 or 8-10</div>}
        {durationInvalid && (
          <div className="helper error">Use Go duration like 45s or 1m30s</div>
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
          onClick={() => setSoundOpen(!soundOpen)}
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
                setSoundOpen(false);
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
                  setSoundOpen(false);
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
