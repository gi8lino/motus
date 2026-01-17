import { useEffect } from "react";

type ModalProps = {
  open: boolean;
  onClose?: () => void;
  children: React.ReactNode;
  closeOnOverlayClick?: boolean;
  closeOnEscape?: boolean;
};

// Modal renders a shared overlay + content shell with optional escape handling.
export function Modal({
  open,
  onClose,
  children,
  closeOnOverlayClick = true,
  closeOnEscape = true,
}: ModalProps) {
  useEffect(() => {
    if (!open || !closeOnEscape || !onClose) return;
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open, closeOnEscape, onClose]);

  if (!open) return null;

  return (
    <div
      className="modal-overlay"
      onClick={closeOnOverlayClick && onClose ? onClose : undefined}
    >
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        {children}
      </div>
    </div>
  );
}
