import { useState } from "react";

import type {
  SessionHistoryItem,
  SessionState,
  SessionStepLog,
  Workout,
} from "../types";
import { formatExerciseLine, formatMillis } from "../utils/format";
import { buildSummary } from "../utils/summary";
import { expandWorkoutSteps } from "../utils/workout";
import { AISummary, HistoryList } from "./HistoryCard";

// HistoryView lists logged sessions and opens a session preview.
export function HistoryView({
  items,
  activeSession,
  onResume,
  loadWorkout,
  onCopySummary,
}: {
  items: SessionHistoryItem[];
  activeSession: SessionState | null;
  onResume: () => void;
  loadWorkout: (id: string) => Promise<Workout>;
  onCopySummary: () => void;
}) {
  const [preview, setPreview] = useState<SessionHistoryItem | null>(null);
  const [previewWorkout, setPreviewWorkout] = useState<Workout | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewDurations, setPreviewDurations] = useState<
    Record<string, number>
  >({});

  // mapHistoryDurations builds a lookup table for step durations.
  const mapHistoryDurations = (steps: SessionStepLog[]) => {
    const map: Record<string, number> = {};
    steps.forEach((s) => {
      const key = `order-${s.stepOrder}`;
      map[key] = s.elapsedMillis;
    });
    return map;
  };

  // mergeWorkoutDurations injects elapsed times into expanded workout steps.
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

  // handleSelect loads preview details for a selected session.
  const handleSelect = (item: SessionHistoryItem) => {
    setPreview(item);
    setPreviewDurations(
      item.steps?.length ? mapHistoryDurations(item.steps) : {},
    );
    setPreviewLoading(true);
    loadWorkout(item.workoutId)
      .then((workout) => setPreviewWorkout(workout))
      .catch(() => setPreviewWorkout(null))
      .finally(() => setPreviewLoading(false));
  };

  return (
    <>
      <section className="panel">
        <div className="panel-header">
          <div>
            <h3>Session history</h3>
            <p className="muted small">
              Completed sessions for the selected user.
            </p>
          </div>
        </div>
        <HistoryList
          items={items}
          activeSession={activeSession}
          onResume={onResume}
          onSelect={handleSelect}
        />
      </section>
      {/* Preview modal */}
      {preview && (
        <div
          className="modal-overlay"
          onClick={() => {
            setPreview(null);
            setPreviewWorkout(null);
          }}
        >
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>Session Overview</h3>
            <p className="muted small">
              Session ID: {preview.sessionId || preview.id}
            </p>
            <div className="stack">
              <div>
                <div className="label">Workout</div>
                <strong>{preview.workoutName || preview.workoutId}</strong>
              </div>
              <div>
                <div className="label">Started</div>
                <div>
                  {preview.startedAt
                    ? new Date(preview.startedAt).toLocaleString()
                    : "n/a"}
                </div>
              </div>
              <div>
                <div className="label">Finished</div>
                <div>
                  {preview.completedAt
                    ? new Date(preview.completedAt).toLocaleString()
                    : "n/a"}
                </div>
              </div>
              <div>
                <div className="label">Duration</div>
                <div>
                  {preview.startedAt && preview.completedAt
                    ? (() => {
                        const start = new Date(preview.startedAt).getTime();
                        const end = new Date(preview.completedAt).getTime();
                        const delta = Math.max(0, end - start);
                        const mins = Math.floor(delta / 60000);
                        const secs = Math.floor((delta % 60000) / 1000)
                          .toString()
                          .padStart(2, "0");
                        return `${mins}:${secs}`;
                      })()
                    : "n/a"}
                </div>
              </div>
              <div>
                <div className="label">Steps</div>
                {previewLoading && (
                  <div className="muted small">Loading steps…</div>
                )}
                {!previewLoading && previewWorkout && (
                  <>
                    {/* Step list */}
                    <ul className="list compact">
                      {mergeWorkoutDurations(
                        previewWorkout,
                        previewDurations,
                      ).map((step, idx) => (
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
                                  ? ` • actual ${formatMillis(step.elapsedMillis)}`
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
                  </>
                )}
                {!previewLoading && !previewWorkout && (
                  <div className="muted small">No step data available.</div>
                )}
              </div>
            </div>
            <AISummary
              summary={
                previewWorkout
                  ? buildSummary({
                      workoutName: preview.workoutName || previewWorkout.name,
                      workoutId: preview.workoutId,
                      userId: preview.userId,
                      startedAt: preview.startedAt,
                      completedAt: preview.completedAt,
                      steps: mergeWorkoutDurations(
                        previewWorkout,
                        previewDurations,
                      ),
                    })
                  : ""
              }
              loading={previewLoading}
              onCopy={onCopySummary}
            />
            <div className="btn-group" style={{ justifyContent: "flex-end" }}>
              <button
                className="btn primary"
                onClick={() => {
                  setPreview(null);
                  setPreviewWorkout(null);
                }}
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
