import type { CatalogExercise, Exercise, SoundOption } from "../../types";
import { ExerciseSelect } from "./ExerciseSelect";
import { SoundIcon } from "../icons/SoundIcon";
import { TrashIcon } from "../icons/TrashIcon";
import { isGoDuration } from "../../utils/time";
import { UI_TEXT } from "../../utils/uiText";
import { isRepRange } from "../../utils/validation";
import { MESSAGES, toErrorMessage } from "../../utils/messages";
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
  repOnly?: boolean;

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
  repOnly = false,
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
  const selectedCatalogExercise = catalog.find(
    (item) => item.id === ex.exerciseId,
  );
  const hasSides =
    selectedCatalogExercise?.hasSides ||
    ex.side === "left" ||
    ex.side === "right";
  const showDuration = isDurationExercise(kind);

  const amountLabel = showDuration
    ? UI_TEXT.labels.duration
    : UI_TEXT.labels.reps;
  const amountPlaceholder = showDuration
    ? UI_TEXT.placeholders.duration
    : UI_TEXT.placeholders.reps;

  const repsValue = (ex.reps || "").trim();
  const durationValue = (ex.duration || "").trim();
  const soundKey = (ex.soundKey || "").trim();

  const soundLabel =
    sounds.find((sound) => sound.key === soundKey)?.label ||
    UI_TEXT.labels.sound;
  const soundSummary = soundKey ? soundLabel : UI_TEXT.labels.subsetSound;

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
              side: selected.hasSides
                ? ex.side === "right"
                  ? "right"
                  : "left"
                : "not_applicable",
            })
          }
          onClear={() =>
            updateExercise(stepIdx, subsetIdx, exIdx, {
              name: "",
              exerciseId: "",
            })
          }
          onAddNew={async () => {
            const newName = await promptUser(UI_TEXT.prompts.exerciseName);
            if (!newName || !newName.trim()) return;

            try {
              const created = await onCreateExercise(newName.trim());
              updateExercise(stepIdx, subsetIdx, exIdx, {
                name: created.name,
                exerciseId: created.id,
              });
            } catch (err) {
              await notifyUser(
                toErrorMessage(err, MESSAGES.createExerciseFailed),
              );
            }
          }}
        />
      </div>

      {hasSides && (
        <div className="field compact side-field">
          <span className="field-label">Side</span>
          <div className="side-picker" role="group" aria-label="Exercise side">
            {(["left", "right"] as const).map((side) => (
              <button
                key={side}
                type="button"
                className="side-option"
                aria-pressed={(ex.side === "right" ? "right" : "left") === side}
                onClick={() =>
                  updateExercise(stepIdx, subsetIdx, exIdx, { side })
                }
              >
                {side === "left" ? "L" : "R"}
              </button>
            ))}
          </div>
        </div>
      )}

      <div className="field compact exercise-type-field">
        <label>Exercise type</label>
        <select
          value={kind}
          onChange={(e) =>
            updateExercise(stepIdx, subsetIdx, exIdx, {
              type: e.target.value as Exercise["type"],
            })
          }
        >
          <option value={EXERCISE_TYPE_REP}>{UI_TEXT.labels.reps}</option>
          <option value={EXERCISE_TYPE_STOPWATCH} disabled={repOnly}>
            Stopwatch
          </option>
          <option value={EXERCISE_TYPE_COUNTDOWN} disabled={repOnly}>
            Countdown
          </option>
        </select>
        {repOnly ? (
          <div className="helper">Supersets support rep exercises only.</div>
        ) : null}
      </div>

      <div className="field compact exercise-amount-field">
        <label>{amountLabel}</label>
        <input
          value={showDuration ? ex.duration || "" : ex.reps || ""}
          onChange={(e) =>
            updateExercise(
              stepIdx,
              subsetIdx,
              exIdx,
              showDuration
                ? { duration: e.target.value }
                : { reps: e.target.value },
            )
          }
          className={repsInvalid || durationInvalid ? "input-error" : undefined}
          placeholder={amountPlaceholder}
        />
        {repsInvalid && <div className="helper error">Use 8 or 8-10</div>}
        {durationInvalid && (
          <div className="helper error">Use Go duration like 45s or 1m30s</div>
        )}
      </div>

      <div className="field compact exercise-weight-field">
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

      <div className="field action compact sound exercise-sound-field">
        <label>{UI_TEXT.labels.sound}</label>
        {repOnly ? (
          <span className="helper">Uses superset sound</span>
        ) : (
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
            title={`${UI_TEXT.labels.sound}: ${soundSummary}`}
            data-label={soundSummary}
            onClick={() => setSoundOpen(!soundOpen)}
          >
            <SoundIcon />
          </button>
        )}

        {!repOnly && soundOpen && (
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

      <div className="field action compact exercise-delete-field">
        <button
          className="btn icon delete mobile-full"
          type="button"
          onClick={() => removeExercise(stepIdx, subsetIdx, exIdx)}
          title={UI_TEXT.titles.removeExercise}
        >
          <span className="desktop-only">
            <TrashIcon />
          </span>
          <span className="mobile-only">{UI_TEXT.titles.removeExercise}</span>
        </button>
      </div>
    </div>
  );
}
