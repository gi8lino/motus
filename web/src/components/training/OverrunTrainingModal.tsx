import { formatCountdownMillis } from "../../utils/format";
import { Modal } from "../common/Modal";

type TrainingOverrunModalProps = {
  show: boolean;
  countdown: number;
  onPause: () => void;
  onPostpone: () => void;
};

// TrainingOverrunModal prompts when a target duration has elapsed.
export function TrainingOverrunModal({
  show,
  countdown,
  onPause,
  onPostpone,
}: TrainingOverrunModalProps) {
  if (!show) return null;

  return (
    <Modal open onClose={onPause}>
      <h3>Still training?</h3>
      <p className="muted">
        You passed the target. Auto-pause in {formatCountdownMillis(countdown)}.
      </p>
      <div className="btn-group" style={{ justifyContent: "flex-end" }}>
        <button className="btn subtle" onClick={onPostpone}>
          Postpone (+30s)
        </button>
        <button className="btn primary" onClick={onPause}>
          Pause
        </button>
      </div>
    </Modal>
  );
}
