type TrainingFinishModalProps = {
  summary: string | null;
  onClose: () => void;
  onCopySummary: () => void;
};

// copySummary writes the summary to the clipboard when supported.
const copySummary = (summary: string, onCopySummary: () => void) => {
  if (navigator?.clipboard?.writeText) {
    navigator.clipboard.writeText(summary).catch(() => {});
  }
  onCopySummary();
};

// TrainingFinishModal shows the end-of-training summary overlay.
export function TrainingFinishModal({
  summary,
  onClose,
  onCopySummary,
}: TrainingFinishModalProps) {
  if (!summary) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h3>Great job!</h3>
        <p className="muted">Training finished. Copy the summary for AI.</p>
        <textarea
          readOnly
          value={summary}
          style={{ width: "100%", minHeight: "180px" }}
        />
        <div className="btn-group" style={{ justifyContent: "flex-end" }}>
          <button
            className="btn subtle"
            onClick={() => copySummary(summary, onCopySummary)}
          >
            Copy
          </button>
          <button className="btn primary" onClick={onClose}>
            Close
          </button>
        </div>
      </div>
    </div>
  );
}
