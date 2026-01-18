import { formatCountdownMillis } from "../../utils/format";
import { Modal } from "../common/Modal";
import { UI_TEXT } from "../../utils/uiText";

type TrainingOverrunModalProps = {
  show: boolean;
  countdown: number;
  onPause: () => void;
  onPostpone: () => void;
  showHours?: boolean;
};

// TrainingOverrunModal prompts when a target duration has elapsed.
export function TrainingOverrunModal({
  show,
  countdown,
  onPause,
  onPostpone,
  showHours,
}: TrainingOverrunModalProps) {
  if (!show) return null;

  return (
    <Modal open onClose={onPause}>
      <h3>{UI_TEXT.pages.training.overrunTitle}</h3>
      <p className="muted">
        {UI_TEXT.pages.training.overrunMessage}{" "}
        {formatCountdownMillis(countdown, { showHours })}.
      </p>
      <div className="btn-group" style={{ justifyContent: "flex-end" }}>
        <button className="btn subtle" onClick={onPostpone}>
          {UI_TEXT.pages.training.overrunPostpone}
        </button>
        <button className="btn primary" onClick={onPause}>
          {UI_TEXT.actions.pause}
        </button>
      </div>
    </Modal>
  );
}
