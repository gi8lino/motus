import { useState } from "react";
import type { View } from "../../types";
import { UI_TEXT } from "../../utils/uiText";

type NavTabsProps = {
  view: View;
  views: View[];
  onSelect: (next: View) => void;
};

const LABELS: Record<View, string> = {
  login: UI_TEXT.nav.login,
  train: UI_TEXT.nav.train,
  workouts: UI_TEXT.nav.workouts,
  templates: UI_TEXT.nav.templates,
  exercises: UI_TEXT.nav.exercises,
  history: UI_TEXT.nav.history,
  profile: UI_TEXT.nav.profile,
  admin: UI_TEXT.nav.admin,
};

/**
 * NavTabs renders the main navigation (desktop + mobile).
 * It is intentionally dumb:
 * - no routing logic
 * - no auth logic
 * - no side effects
 */
export function NavTabs({ view, views, onSelect }: NavTabsProps) {
  const [open, setOpen] = useState(false);

  return (
    <div className="nav-shell">
      {/* Mobile toggle */}
      <button
        className="btn icon nav-toggle"
        type="button"
        aria-label={UI_TEXT.accessibility.navToggle}
        onClick={() => setOpen((v) => !v)}
      >
        ☰
      </button>

      <nav className={open ? "nav-menu open" : "nav-menu"}>
        {views.map((v) => (
          <button
            key={v}
            type="button"
            className={view === v ? "tab active" : "tab"}
            onClick={() => {
              onSelect(v);
              setOpen(false);
            }}
          >
            {LABELS[v]}
          </button>
        ))}
      </nav>
    </div>
  );
}
