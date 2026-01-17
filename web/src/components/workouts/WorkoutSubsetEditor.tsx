import type {
  CatalogExercise,
  Exercise,
  SoundOption,
  WorkoutSubset,
  WorkoutStep,
} from "../../types";
import { TrashIcon } from "../icons/TrashIcon";
import { isGoDuration } from "../../utils/time";
import { WorkoutExerciseRow } from "./WorkoutExerciseRow";
import { UI_TEXT } from "../../utils/uiText";

type MutableRef<T> = { current: T };
type DragExercise = { stepIdx: number; subsetIdx: number; idx: number };

type WorkoutSubsetEditorProps = {
  stepIdx: number;
  step: WorkoutStep;

  subset: WorkoutSubset;
  subsetIdx: number;
  subsetsLength: number;

  sounds: SoundOption[];
  catalog: CatalogExercise[];

  addExercise: (stepIdx: number, subsetIdx: number) => void;
  removeSubset: (stepIdx: number, subsetIdx: number) => void;

  updateSubset: (
    stepIdx: number,
    subsetIdx: number,
    patch: Partial<WorkoutSubset>,
  ) => void;

  updateExercise: (
    stepIdx: number,
    subsetIdx: number,
    exIdx: number,
    patch: Partial<Exercise>,
  ) => void;

  removeExercise: (stepIdx: number, subsetIdx: number, exIdx: number) => void;

  dragExerciseRef: MutableRef<DragExercise | null>;
  moveExercise: (
    stepIdx: number,
    subsetIdx: number,
    from: number,
    to: number,
  ) => void;

  soundPopoverRef: React.RefObject<HTMLDivElement | null>;
  isSoundOpen: (stepIdx: number, subsetIdx: number, exIdx: number) => boolean;
  setSoundOpen: (
    stepIdx: number,
    subsetIdx: number,
    exIdx: number,
    open: boolean,
  ) => void;

  promptUser: (
    message: string,
    defaultValue?: string,
  ) => Promise<string | null>;
  onCreateExercise: (name: string) => Promise<CatalogExercise>;
  notifyUser: (message: string) => Promise<void>;
};

export function WorkoutSubsetEditor({
  stepIdx,
  step,
  subset,
  subsetIdx,
  subsetsLength,
  sounds,
  catalog,
  addExercise,
  removeSubset,
  updateSubset,
  updateExercise,
  removeExercise,
  dragExerciseRef,
  moveExercise,
  soundPopoverRef,
  isSoundOpen,
  setSoundOpen,
  promptUser,
  onCreateExercise,
  notifyUser,
}: WorkoutSubsetEditorProps) {
  const hasMultiple = subsetsLength > 1;
  const subsetLabel = subset.name?.trim() || `Subset ${subsetIdx + 1}`;
  const subsetExercises = subset.exercises || [];

  const targetDurationValue = (subset.duration || "").trim();
  const targetDurationInvalid =
    targetDurationValue !== "" && !isGoDuration(targetDurationValue);

  const isSuperset = Boolean(subset.superset);

  return (
    <div
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
            <label
              className="switch superset-toggle"
              title={UI_TEXT.titles.supersetTooltip}
            >
              <input
                type="checkbox"
                checked={isSuperset}
                onChange={(event) =>
                  updateSubset(stepIdx, subsetIdx, {
                    superset: event.target.checked,
                  })
                }
              />
              <span className="switch-slider" aria-hidden="true" />
              <span className="switch-label">Superset</span>
            </label>

            <button
              className="btn icon delete icon-only"
              type="button"
              onClick={() => removeSubset(stepIdx, subsetIdx)}
              disabled={subsetsLength <= 1}
              title={UI_TEXT.titles.removeSubset}
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
            updateSubset(stepIdx, subsetIdx, { name: event.target.value })
          }
        />
      </div>

      <div className="field">
        <label>Target time (optional, e.g. 45s)</label>
        <input
          value={subset.duration || ""}
          onChange={(event) =>
            updateSubset(stepIdx, subsetIdx, { duration: event.target.value })
          }
          className={targetDurationInvalid ? "input-error" : undefined}
        />
        {targetDurationInvalid && (
          <div className="helper error">Use Go duration like 45s or 1m30s</div>
        )}
      </div>

      <div className="field">
        <label>Sound</label>
        <select
          value={subset.soundKey || ""}
          onChange={(event) =>
            updateSubset(stepIdx, subsetIdx, { soundKey: event.target.value })
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

      {!hasMultiple && (
        <div className="subset-actions superset-toggle-row">
          <label
            className="switch superset-toggle"
            title={UI_TEXT.titles.supersetTooltip}
          >
            <input
              type="checkbox"
              checked={isSuperset}
              onChange={(event) =>
                updateSubset(stepIdx, subsetIdx, {
                  superset: event.target.checked,
                })
              }
            />
            <span className="switch-slider" aria-hidden="true" />
            <span className="switch-label">Superset</span>
          </label>
        </div>
      )}

      {subsetExercises.length ? (
        subsetExercises.map((exercise, exIdx) => (
          <WorkoutExerciseRow
            key={`${subset.id || `${stepIdx}-${subsetIdx}`}-${exIdx}`}
            trainingKey={`${step.id || "step"}:${subset.id || "subset"}`}
            stepIdx={stepIdx}
            subsetIdx={subsetIdx}
            exIdx={exIdx}
            ex={exercise}
            catalog={catalog}
            sounds={sounds}
            dragExerciseRef={dragExerciseRef}
            moveExercise={moveExercise}
            updateExercise={updateExercise}
            removeExercise={removeExercise}
            promptUser={promptUser}
            onCreateExercise={onCreateExercise}
            notifyUser={notifyUser}
            soundPopoverRef={soundPopoverRef}
            soundOpen={isSoundOpen(stepIdx, subsetIdx, exIdx)}
            setSoundOpen={(open) =>
              setSoundOpen(stepIdx, subsetIdx, exIdx, open)
            }
          />
        ))
      ) : (
        <div className="muted small">
          Add at least one exercise for this subset.
        </div>
      )}

      <div className="btn-group">
        <button
          className="btn outline"
          type="button"
          onClick={() => addExercise(stepIdx, subsetIdx)}
        >
          Add Exercise
        </button>
      </div>
    </div>
  );
}
