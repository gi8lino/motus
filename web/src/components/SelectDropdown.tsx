import { useEffect, useRef, useState } from "react";

type SelectItem = {
  id: string;
  label: string;
  right?: React.ReactNode;
};

// SelectDropdown renders a generic dropdown with custom option content.
export function SelectDropdown({
  items,
  value,
  placeholder,
  onSelect,
  onClear,
  renderRight,
  renderSelectedRight,
  addLabel,
  onAddNew,
}: {
  items: SelectItem[];
  value: string | null;
  placeholder: string;
  onSelect: (item: SelectItem) => void;
  onClear?: () => void;
  renderRight?: (item: SelectItem) => React.ReactNode;
  renderSelectedRight?: (item: SelectItem | null) => React.ReactNode;
  addLabel?: string;
  onAddNew?: () => void | Promise<void>;
}) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement | null>(null);

  const selected: SelectItem | null = value
    ? (items.find((item) => item.id === value) ?? null)
    : null;

  useEffect(() => {
    const handleClick = (event: MouseEvent) => {
      if (!rootRef.current) return;
      if (rootRef.current.contains(event.target as Node)) return;
      setOpen(false);
    };
    if (open) {
      document.addEventListener("mousedown", handleClick);
    }
    return () => document.removeEventListener("mousedown", handleClick);
  }, [open]);

  return (
    <div className="select-dropdown" ref={rootRef}>
      <button
        className="select-dropdown-trigger"
        type="button"
        onClick={() => setOpen((prev) => !prev)}
      >
        <span className={`select-dropdown-label ${selected ? "" : "muted"}`}>
          {selected ? selected.label : placeholder}
        </span>
        {renderSelectedRight?.(selected)}
        <span className="select-dropdown-chevron">{open ? "▼" : "▶"}</span>
      </button>
      {open && (
        <div className="select-dropdown-menu">
          {selected && onClear && (
            <button
              className="select-dropdown-option muted"
              type="button"
              onClick={() => {
                onClear();
                setOpen(false);
              }}
            >
              Clear selection
            </button>
          )}
          {items.map((item) => (
            <button
              key={item.id}
              className="select-dropdown-option"
              type="button"
              onClick={() => {
                onSelect(item);
                setOpen(false);
              }}
            >
              <span>{item.label}</span>
              {renderRight?.(item) ?? item.right}
            </button>
          ))}
          {addLabel && onAddNew && (
            <button
              className="select-dropdown-option add"
              type="button"
              onClick={async () => {
                await onAddNew();
                setOpen(false);
              }}
            >
              {addLabel}
            </button>
          )}
        </div>
      )}
    </div>
  );
}
