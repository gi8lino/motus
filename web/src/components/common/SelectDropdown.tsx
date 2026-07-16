import { useEffect, useRef, useState } from "react";
import { matchesSearch } from "../../utils/search";

type SelectItem = {
  id: string;
  label: string;
  searchText?: string;
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
  searchable = false,
}: {
  items: SelectItem[];
  value: string | null;
  placeholder: string;
  onSelect: (item: SelectItem) => void;
  onClear?: () => void;
  renderRight?: (item: SelectItem) => React.ReactNode;
  renderSelectedRight?: (
    item: SelectItem | null | undefined,
  ) => React.ReactNode;
  addLabel?: string;
  onAddNew?: () => void | Promise<void>;
  searchable?: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const rootRef = useRef<HTMLDivElement | null>(null);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const typeBufferRef = useRef("");
  const typeTimeoutRef = useRef<number | null>(null);

  const selected: SelectItem | null = value
    ? (items.find((item) => item.id === value) ?? null)
    : null;
  const normalizedQuery = query.trim().toLowerCase();
  const visibleItems = normalizedQuery
    ? items.filter((item) => matchesSearch(item, normalizedQuery))
    : items;

  // Close the menu when the user clicks outside the dropdown.
  useEffect(() => {
    // handleClick closes the menu on outside clicks.
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

  const resetTypeAhead = () => {
    typeBufferRef.current = "";
    if (typeTimeoutRef.current) {
      window.clearTimeout(typeTimeoutRef.current);
      typeTimeoutRef.current = null;
    }
  };

  const moveToMatch = (term: string) => {
    if (!menuRef.current) return false;
    const options = Array.from(
      menuRef.current.querySelectorAll<HTMLButtonElement>(
        ".select-dropdown-option",
      ),
    ).filter((button) => !button.classList.contains("muted"));
    const match = options.find((button) =>
      button.textContent?.toLowerCase().startsWith(term),
    );
    if (match) {
      match.focus();
      return true;
    }
    return false;
  };

  // handleTypeAhead focuses the first matching option by typed prefix.
  const handleTypeAhead = (event: React.KeyboardEvent) => {
    if (!open) return;
    if (event.target instanceof HTMLInputElement) return;
    if (event.metaKey || event.ctrlKey || event.altKey) return;
    const key = event.key;
    if (key === "Enter") {
      const target = event.target as HTMLElement | null;
      const button = target?.closest<HTMLButtonElement>(
        ".select-dropdown-option",
      );
      if (button && menuRef.current?.contains(button)) {
        button.click();
        event.preventDefault();
      }
      return;
    }
    if (key.length !== 1) return;
    const next = `${typeBufferRef.current}${key}`.toLowerCase();
    typeBufferRef.current = next;
    moveToMatch(next);
    if (typeTimeoutRef.current) {
      window.clearTimeout(typeTimeoutRef.current);
    }
    typeTimeoutRef.current = window.setTimeout(resetTypeAhead, 700);
  };

  // Reset the type-ahead buffer when the menu closes.
  useEffect(() => {
    if (!open) {
      resetTypeAhead();
      setQuery("");
    }
  }, [open]);

  // Clear any pending timeouts on unmount.
  useEffect(() => {
    return () => {
      if (typeTimeoutRef.current) {
        window.clearTimeout(typeTimeoutRef.current);
      }
    };
  }, []);

  return (
    <div className="select-dropdown" ref={rootRef}>
      <button
        className="select-dropdown-trigger"
        type="button"
        onClick={() => setOpen((prev) => !prev)}
        onKeyDown={handleTypeAhead}
      >
        <span className={`select-dropdown-label ${selected ? "" : "muted"}`}>
          {selected ? selected.label : placeholder}
        </span>
        {renderSelectedRight?.(selected ?? null)}
        <span className="select-dropdown-chevron">{open ? "▼" : "▶"}</span>
      </button>
      {open && (
        <div
          className="select-dropdown-menu"
          ref={menuRef}
          onKeyDown={handleTypeAhead}
        >
          {searchable ? (
            <input
              className="select-dropdown-search"
              type="search"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search exercises or labels…"
              aria-label="Search exercises"
              autoFocus
            />
          ) : null}
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
          {visibleItems.map((item) => (
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
          {searchable && visibleItems.length === 0 ? (
            <div className="select-dropdown-empty">No matching exercises</div>
          ) : null}
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
