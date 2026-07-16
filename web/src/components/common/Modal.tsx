import { useEffect, useRef } from "react";

type ModalProps = {
  open: boolean;
  onClose?: () => void | Promise<void>;
  children: React.ReactNode;
  closeOnOverlayClick?: boolean;
  closeOnEscape?: boolean;
  ariaLabel?: string;
};

// Modal renders a shared overlay + content shell with optional escape handling.
export function Modal({
  open,
  onClose,
  children,
  closeOnOverlayClick = true,
  closeOnEscape = true,
  ariaLabel = "Dialog",
}: ModalProps) {
  const modalRef = useRef<HTMLDivElement | null>(null);
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

  useEffect(() => {
    if (!open) return;
    const previouslyFocused = document.activeElement as HTMLElement | null;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    const focusableSelector = [
      "button:not([disabled])",
      "input:not([disabled])",
      "select:not([disabled])",
      "textarea:not([disabled])",
      "a[href]",
      '[tabindex]:not([tabindex="-1"])',
    ].join(",");

    const focusFirst = () => {
      const first =
        modalRef.current?.querySelector<HTMLElement>(focusableSelector);
      (first ?? modalRef.current)?.focus();
    };
    const frame = window.requestAnimationFrame(focusFirst);

    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape" && closeOnEscape && onCloseRef.current) {
        event.preventDefault();
        void onCloseRef.current();
        return;
      }
      if (event.key !== "Tab" || !modalRef.current) return;

      const focusable = Array.from(
        modalRef.current.querySelectorAll<HTMLElement>(focusableSelector),
      );
      if (!focusable.length) {
        event.preventDefault();
        modalRef.current.focus();
        return;
      }

      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    };
    window.addEventListener("keydown", handler);
    return () => {
      window.cancelAnimationFrame(frame);
      window.removeEventListener("keydown", handler);
      document.body.style.overflow = previousOverflow;
      previouslyFocused?.focus();
    };
  }, [open, closeOnEscape]);

  if (!open) return null;

  return (
    <div
      className="modal-overlay"
      onClick={
        closeOnOverlayClick && onCloseRef.current
          ? () => {
              void onCloseRef.current?.();
            }
          : undefined
      }
    >
      <div
        ref={modalRef}
        className="modal"
        role="dialog"
        aria-modal="true"
        aria-label={ariaLabel}
        tabIndex={-1}
        onClick={(e) => e.stopPropagation()}
      >
        {children}
      </div>
    </div>
  );
}
