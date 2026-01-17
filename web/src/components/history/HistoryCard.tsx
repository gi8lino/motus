import type { TrainingHistoryItem, TrainingState } from "../../types";

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
    return <p className="muted">No history yet.</p>;
  return (
    <ul className="list">
      {/* Active training banner */}
      {hasActive && (
        <li className="list-item">
          <div className="list-row">
            <div>
              <strong>Active training</strong>
              <div className="muted small">
                {activeTraining?.workoutName || activeTraining?.workoutId} • in
                progress
              </div>
            </div>
            {onResume && (
              <button className="btn primary" onClick={onResume}>
                Resume
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
                  : "Not started"}{" "}
                •{" "}
                {item.completedAt
                  ? `Finished ${new Date(item.completedAt).toLocaleString()}`
                  : "Not finished"}
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
      <div className="label">AI-ready summary</div>
      {loading && <div className="muted small">Loading steps…</div>}
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
          Copy summary
        </button>
      </div>
    </div>
  );
}
