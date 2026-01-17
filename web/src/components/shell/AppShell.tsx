import type { ReactNode } from "react";
import type { User, View } from "../../types";
import { PROMPTS } from "../../utils/messages";
import { BrandHeader } from "../common/BrandHeader";
import { NavTabs } from "./NavigationTabs";

type AppShellProps = {
  view: View;
  onViewChange: (view: View) => void;

  currentUser: User | null;
  authHeaderEnabled: boolean;

  onLogout?: () => void;

  // Resume toast
  resumeOpen: boolean;
  resumeText: string;
  onResume: () => void;
  onDismissResume: () => void;

  // Global toast
  toast: string | null;

  // Footer
  appVersion: string;

  children: ReactNode;
};

export function AppShell({
  view,
  onViewChange,
  currentUser,
  authHeaderEnabled,
  onLogout,
  resumeOpen,
  resumeText,
  onResume,
  onDismissResume,
  toast,
  appVersion,
  children,
}: AppShellProps) {
  const isAuthed = Boolean(authHeaderEnabled || currentUser);

  const availableViews: View[] = (
    [
      "train",
      "workouts",
      "templates",
      "exercises",
      "history",
      "profile",
      "admin",
    ] as View[]
  ).filter((v) => (v === "admin" ? Boolean(currentUser?.isAdmin) : true));

  return (
    <>
      <div className="shell">
        {resumeOpen && (
          <div className="toast">
            <div>
              <strong>{PROMPTS.resumeTrainingTitle}</strong>
              <div className="muted small">{resumeText}</div>
            </div>
            <div className="btn-group">
              <button className="btn primary" onClick={onResume} type="button">
                {PROMPTS.resumeTrainingConfirm}
              </button>
              <button
                className="btn subtle"
                onClick={onDismissResume}
                type="button"
              >
                Dismiss
              </button>
            </div>
          </div>
        )}

        <header className="topbar">
          <BrandHeader />

          {isAuthed && (
            <div className="topbar-actions">
              <NavTabs
                view={view}
                views={availableViews}
                onSelect={onViewChange}
              />

              {!authHeaderEnabled && currentUser && onLogout && (
                <button className="btn subtle" onClick={onLogout} type="button">
                  Logout
                </button>
              )}
            </div>
          )}
        </header>

        {/* 🔥 CRITICAL FIX: key forces full remount per view */}
        <main key={view}>{children}</main>
      </div>

      {toast && <div className="toast-floating">{toast}</div>}

      <footer className="app-footer">
        <div className="app-footer-inner">
          <span>© 2025 Motus | Version: {appVersion}</span>
        </div>
      </footer>
    </>
  );
}
