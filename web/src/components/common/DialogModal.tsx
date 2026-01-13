type DialogType = "alert" | "confirm" | "prompt";

export type DialogState = {
  type: DialogType;
  message: string;
  title?: string;
  defaultValue?: string;
  placeholder?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  resolve: (value: any) => void;
};

type DialogModalProps = {
  dialog: DialogState;
  value: string;
  onValueChange: (value: string) => void;
  onClose: () => void;
};

// resolveDialog sends the confirm action for the dialog type.
const resolveDialog = (dialog: DialogState, value: string) => {
  if (dialog.type === "confirm") {
    dialog.resolve(true);
    return;
  }
  if (dialog.type === "prompt") {
    dialog.resolve(value);
    return;
  }
  dialog.resolve(undefined);
};

// dismissDialog sends the cancel action for the dialog type.
const dismissDialog = (dialog: DialogState) => {
  if (dialog.type === "confirm") {
    dialog.resolve(false);
    return;
  }
  if (dialog.type === "prompt") {
    dialog.resolve(null);
    return;
  }
  dialog.resolve(undefined);
};

// dialogTitle resolves the dialog heading text.
const dialogTitle = (dialog: DialogState) => {
  if (dialog.title) return dialog.title;
  if (dialog.type === "confirm") return "Confirm";
  return "Message";
};

// confirmLabel resolves the confirmation button label.
const confirmLabel = (dialog: DialogState) => {
  if (dialog.confirmLabel) return dialog.confirmLabel;
  if (dialog.type === "confirm") return "Confirm";
  if (dialog.type === "prompt") return "Save";
  return "OK";
};

// DialogModal renders the shared alert/confirm/prompt overlay.
export default function DialogModal({
  dialog,
  value,
  onValueChange,
  onClose,
}: DialogModalProps) {
  return (
    <div
      className="modal-overlay"
      onClick={() => {
        dismissDialog(dialog);
        onClose();
      }}
    >
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <h3>{dialogTitle(dialog)}</h3>
        <p className="muted">{dialog.message}</p>
        {dialog.type === "prompt" && (
          <input
            autoFocus
            value={value}
            placeholder={dialog.placeholder || ""}
            onChange={(e) => onValueChange(e.target.value)}
          />
        )}
        <div className="btn-group" style={{ justifyContent: "flex-end" }}>
          {dialog.type !== "alert" && (
            <button
              className="btn subtle"
              onClick={() => {
                dismissDialog(dialog);
                onClose();
              }}
            >
              {dialog.cancelLabel || "Cancel"}
            </button>
          )}
          <button
            className="btn primary"
            onClick={() => {
              resolveDialog(dialog, value);
              onClose();
            }}
          >
            {confirmLabel(dialog)}
          </button>
        </div>
      </div>
    </div>
  );
}
