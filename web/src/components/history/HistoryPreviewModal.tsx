import type {
  TrainingHistoryItem,
  TrainingStepLog,
  Workout,
} from "../../types";
import { formatExerciseLine, formatElapsedMillis } from "../../utils/format";
import { buildSummary } from "../../utils/summary";
import { AISummary } from "./HistoryCard";
import { Modal } from "../common/Modal";

type HistoryPreviewModalProps = {
  preview: TrainingHistoryItem | null;
  workout: Workout | null;
  loading: boolean;
  onClose: () => void;
  onCopySummary: () => void;
};

// historySteps returns the immutable completion snapshot, falling back for legacy logs.
function historySteps(
  preview: TrainingHistoryItem,
  workout: Workout | null,
): TrainingStepLog[] {
  if (preview.steps?.length) return preview.steps;
  return (workout?.steps || []).map((step, index) => ({
    id: step.id || `legacy-step-${index}`,
    trainingId: preview.trainingId || preview.id,
    stepOrder: index,
    type: step.type,
    name: step.name,
    estimatedSeconds: step.estimatedSeconds || 0,
    elapsedMillis: 0,
    exercises: step.exercises,
  }));
}

// formatTimestamp renders an ISO timestamp into a readable string.
const formatTimestamp = (value?: string) => {
  if (!value) return "n/a";
  const date = new Date(value);
  if (Number.isNaN(date.valueOf())) return "n/a";
  return date.toLocaleString();
};

// formatDuration renders an elapsed duration between two timestamps.
const formatDuration = (startedAt?: string, completedAt?: string) => {
  if (!startedAt || !completedAt) return "n/a";
  const start = new Date(startedAt).getTime();
  const end = new Date(completedAt).getTime();
  const delta = Math.max(0, end - start);
  const mins = Math.floor(delta / 60000);
  const secs = Math.floor((delta % 60000) / 1000)
    .toString()
    .padStart(2, "0");
  return `${mins}:${secs}`;
};

// HistoryPreviewModal shows a training preview overlay.
export function HistoryPreviewModal({
  preview,
  workout,
  loading,
  onClose,
  onCopySummary,
}: HistoryPreviewModalProps) {
  if (!preview) return null;
  const completedSteps = historySteps(preview, workout);

  return (
    <Modal open onClose={onClose}>
      <h3>Training overview</h3>
      <p className="muted small">
        Training ID: {preview.trainingId || preview.id}
      </p>
      <div className="stack">
        <div>
          <div className="label">Workout</div>
          <strong>{preview.workoutName || preview.workoutId}</strong>
        </div>
        <div>
          <div className="label">Started</div>
          <div>{formatTimestamp(preview.startedAt)}</div>
        </div>
        <div>
          <div className="label">Finished</div>
          <div>{formatTimestamp(preview.completedAt)}</div>
        </div>
        <div>
          <div className="label">Duration</div>
          <div>{formatDuration(preview.startedAt, preview.completedAt)}</div>
        </div>
        <div>
          <div className="label">Steps</div>
          {loading && <div className="muted small">Loading steps…</div>}
          {!loading && completedSteps.length > 0 && (
            <ul className="list compact">
              {completedSteps.map((step, idx) => (
                <li key={step.id || idx} className="list-item">
                  <div className="list-row">
                    <div>
                      <strong>
                        {idx + 1}. {step.name}
                      </strong>
                      <div className="muted small">
                        {step.type} •{" "}
                        {step.estimatedSeconds
                          ? `target ${step.estimatedSeconds}s`
                          : "open"}
                        {step.elapsedMillis
                          ? ` • actual ${formatElapsedMillis(step.elapsedMillis)}`
                          : ""}
                      </div>
                      {step.exercises?.length ? (
                        <div className="muted small">
                          {step.exercises
                            .map((ex) => formatExerciseLine(ex))
                            .filter(Boolean)
                            .join(" | ")}
                        </div>
                      ) : null}
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
          {!loading && completedSteps.length === 0 && (
            <div className="muted small">No step data available.</div>
          )}
        </div>
      </div>
      <AISummary
        summary={
          completedSteps.length
            ? buildSummary({
                workoutName: preview.workoutName || workout?.name,
                workoutId: preview.workoutId,
                userId: preview.userId,
                startedAt: preview.startedAt,
                completedAt: preview.completedAt,
                steps: completedSteps,
              })
            : ""
        }
        loading={loading}
        onCopy={onCopySummary}
      />
      <div className="btn-group" style={{ justifyContent: "flex-end" }}>
        <button className="btn primary" onClick={onClose}>
          Close
        </button>
      </div>
    </Modal>
  );
}
