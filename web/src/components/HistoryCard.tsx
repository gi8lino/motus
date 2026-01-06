import type { SessionHistoryItem, SessionState } from "../types";

// HistoryList renders past sessions and supports selection/resume.
export function HistoryList({
  items,
  activeSession,
  onResume,
  onSelect,
}: {
  items: SessionHistoryItem[];
  activeSession?: SessionState | null;
  onResume?: () => void;
  onSelect?: (item: SessionHistoryItem) => void;
}) {
  const hasActive = activeSession && !activeSession.done;
  if (!items.length && !hasActive)
    return <p className="muted">No history yet.</p>;
  return (
    <ul className="list">
      {/* Active session banner */}
      {hasActive && (
        <li className="list-item">
          <div className="list-row">
            <div>
              <strong>Active training</strong>
              <div className="muted small">
                {activeSession?.workoutName || activeSession?.workoutId} • in
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
