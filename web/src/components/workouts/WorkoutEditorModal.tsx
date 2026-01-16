import { WorkoutForm, type WorkoutFormProps } from "./WorkoutForm";

type WorkoutFormModalProps = WorkoutFormProps & {
  open: boolean;
  onClose: () => void;
};

// WorkoutEditorModal wraps WorkoutForm in the shared modal shell.
export function WorkoutEditorModal({
  open,
  onClose,
  ...formProps
}: WorkoutFormModalProps) {
  if (!open) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <WorkoutForm {...formProps} onClose={onClose} />
      </div>
    </div>
  );
}
