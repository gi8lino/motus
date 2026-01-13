import { formatMillis } from "../../utils/format";

type SessionOverrunModalProps = {
  show: boolean;
  countdown: number;
  onPause: () => void;
  onPostpone: () => void;
};

// SessionOverrunModal prompts when a target duration has elapsed.
export function SessionOverrunModal({
  show,
  countdown,
  onPause,
  onPostpone,
}: SessionOverrunModalProps) {
  if (!show) return null;

  return (
    <div className="modal-overlay" onClick={onPause}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h3>Still training?</h3>
        <p className="muted">
          You passed the target. Auto-pause in {formatMillis(countdown)}.
        </p>
        <div className="btn-group" style={{ justifyContent: "flex-end" }}>
          <button className="btn subtle" onClick={onPostpone}>
            Postpone (+30s)
          </button>
          <button className="btn primary" onClick={onPause}>
            Pause
          </button>
        </div>
      </div>
    </div>
  );
}
