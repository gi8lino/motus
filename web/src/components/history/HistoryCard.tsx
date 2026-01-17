import type { TrainingHistoryItem, TrainingState } from "../../types";
import { UI_TEXT } from "../../utils/uiText";

// HistoryList renders past trainings and supports selection/resume.
export function HistoryList({
  items,
  activeTraining,
  onResume,
  onSelect,
}: {
  items: TrainingHistoryItem[];
  activeTraining?: TrainingState | null;
  onResume?: () => void;
  onSelect?: (item: TrainingHistoryItem) => void;
}) {
  const hasActive = activeTraining && !activeTraining.done;
  if (!items.length && !hasActive)
    return <p className="muted">{UI_TEXT.history.noHistory}</p>;
  return (
    <ul className="list">
      {/* Active training banner */}
      {hasActive && (
        <li className="list-item">
          <div className="list-row">
            <div>
              <strong>{UI_TEXT.history.activeTraining}</strong>
              <div className="muted small">
                {activeTraining?.workoutName || activeTraining?.workoutId} •{" "}
                {UI_TEXT.history.inProgress}
              </div>
            </div>
            {onResume && (
              <button className="btn primary" onClick={onResume}>
                {UI_TEXT.history.resume}
              </button>
            )}
          </div>
        </li>
      )}
      {items.map((item) => (
        <li
          key={item.id}
          className="list-item"
          onClick={() => onSelect?.(item)}
          style={{ cursor: onSelect ? "pointer" : "default" }}
        >
          <div className="list-row">
            <div>
              <strong>{item.workoutName || item.workoutId}</strong>
              <div className="muted small">
                {item.startedAt
                  ? new Date(item.startedAt).toLocaleString()
                  : UI_TEXT.history.notStarted}{" "}
                •{" "}
                {item.completedAt
                  ? `${UI_TEXT.history.finishedPrefix} ${new Date(item.completedAt).toLocaleString()}`
                  : UI_TEXT.history.notFinished}
              </div>
            </div>
          </div>
        </li>
      ))}
    </ul>
  );
}

// AISummary builds a copyable summary string for AI tools.
export function AISummary({
  summary,
  loading,
  onCopy,
}: {
  summary?: string;
  loading: boolean;
  onCopy: () => void;
}) {
  const fallbackSummary = summary || "";
  return (
    <div className="stack">
      <div className="label">{UI_TEXT.history.aiSummary}</div>
      {loading && (
        <div className="muted small">{UI_TEXT.history.loadingSteps}</div>
      )}
      <textarea
        readOnly
        value={fallbackSummary}
        style={{ width: "100%", minHeight: "180px" }}
      />
      <div className="btn-group" style={{ justifyContent: "flex-end" }}>
        <button
          className="btn subtle"
          type="button"
          onClick={() => {
            if (navigator?.clipboard?.writeText) {
              navigator.clipboard.writeText(summary || "").catch(() => {});
            }
            onCopy();
          }}
        >
          {UI_TEXT.history.copySummary}
        </button>
      </div>
    </div>
  );
}
