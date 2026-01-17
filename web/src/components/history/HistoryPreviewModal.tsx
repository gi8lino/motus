import { useMemo } from "react";

import type {
  TrainingHistoryItem,
  TrainingStepLog,
  Workout,
} from "../../types";
import { formatExerciseLine, formatElapsedMillis } from "../../utils/format";
import { buildSummary } from "../../utils/summary";
import { expandWorkoutSteps } from "../../utils/workout";
import { AISummary } from "./HistoryCard";
import { Modal } from "../common/Modal";

type HistoryPreviewModalProps = {
  preview: TrainingHistoryItem | null;
  workout: Workout | null;
  loading: boolean;
  onClose: () => void;
  onCopySummary: () => void;
};

// mapHistoryDurations builds an elapsed duration lookup by step order.
const mapHistoryDurations = (steps: TrainingStepLog[]) => {
  const map: Record<string, number> = {};
  steps.forEach((s) => {
    const key = `order-${s.stepOrder}`;
    map[key] = s.elapsedMillis;
  });
  return map;
};

// mergeWorkoutDurations merges logged timings into expanded workout steps.
const mergeWorkoutDurations = (
  workout: Workout,
  durations: Record<string, number>,
) => {
  const expanded = expandWorkoutSteps(workout.steps || []);
  return expanded.map((step, idx) => ({
    ...step,
    elapsedMillis:
      durations[`order-${idx}`] ?? durations[step.id || `step-${idx}`] ?? 0,
  }));
};

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
  const previewDurations = useMemo(
    () => (preview?.steps?.length ? mapHistoryDurations(preview.steps) : {}),
    [preview],
  );

  const expandedSteps = useMemo(() => {
    if (!workout) return [];
    return mergeWorkoutDurations(workout, previewDurations);
  }, [previewDurations, workout]);

  if (!preview) return null;

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
          {!loading && workout && (
            <ul className="list compact">
              {expandedSteps.map((step, idx) => (
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
                          : step.duration || "open"}
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
          {!loading && !workout && (
            <div className="muted small">No step data available.</div>
          )}
        </div>
      </div>
      <AISummary
        summary={
          workout
            ? buildSummary({
                workoutName: preview.workoutName || workout.name,
                workoutId: preview.workoutId,
                userId: preview.userId,
                startedAt: preview.startedAt,
                completedAt: preview.completedAt,
                steps: expandedSteps,
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
