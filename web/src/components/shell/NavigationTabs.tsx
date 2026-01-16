import { useState } from "react";
import type { View } from "../../types";

type NavTabsProps = {
  view: View;
  views: View[];
  onSelect: (next: View) => void;
};

const LABELS: Record<View, string> = {
  login: "Login",
  train: "Train",
  workouts: "Workouts",
  templates: "Templates",
  exercises: "Exercises",
  history: "History",
  profile: "Profile",
  admin: "Admin",
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
        aria-label="Toggle navigation"
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
