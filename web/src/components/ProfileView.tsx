import { useEffect, useState } from "react";
import type { RefObject } from "react";
import type { Workout } from "../types";

type ProfileTab = "settings" | "password" | "transfer";
type ThemeMode = "auto" | "dark" | "light";

// ProfileView renders account preferences and transfer actions.
export function ProfileView({
  profileTab,
  onProfileTabChange,
  themeMode,
  onThemeChange,
  repeatRestAfterLastDefault,
  onRepeatRestAfterLastDefaultChange,
  exportWorkoutId,
  onExportWorkoutChange,
  activeWorkouts,
  onExportWorkout,
  onImportWorkout,
  onPasswordChange,
  importInputRef,
  authHeaderEnabled,
}: {
  profileTab: ProfileTab;
  onProfileTabChange: (tab: ProfileTab) => void;
  themeMode: ThemeMode;
  onThemeChange: (mode: ThemeMode) => void;
  repeatRestAfterLastDefault: boolean;
  onRepeatRestAfterLastDefaultChange: (value: boolean) => void;
  exportWorkoutId: string;
  onExportWorkoutChange: (id: string) => void;
  activeWorkouts: Workout[];
  onExportWorkout: () => void | Promise<void>;
  onImportWorkout: (file: File) => void | Promise<void>;
  onPasswordChange: (
    currentPassword: string,
    newPassword: string,
  ) => void | Promise<void>;
  importInputRef: RefObject<HTMLInputElement | null>;
  authHeaderEnabled: boolean;
}) {
  const canExport = Boolean(exportWorkoutId);
  useEffect(() => {
    if (authHeaderEnabled && profileTab === "password") {
      onProfileTabChange("settings");
    }
  }, [authHeaderEnabled, profileTab, onProfileTabChange]);
  return (
    <section className="panel">
      <div className="panel-header">
        <div>
          <h3>Profile</h3>
          <p className="muted small">Manage your local account preferences.</p>
        </div>
      </div>
      <div className="profile-layout">
        <div className="profile-content">
          {/* Tab content */}
          {profileTab === "settings" && (
            <div className="stack">
              <div className="field">
                <label>Theme</label>
                <select
                  value={themeMode}
                  onChange={(e) => onThemeChange(e.target.value as ThemeMode)}
                >
                  <option value="auto">Auto (system)</option>
                  <option value="dark">Dark</option>
                  <option value="light">Light</option>
                </select>
              </div>
              <div className="divider" />
              <label className="field checkbox">
                <input
                  type="checkbox"
                  checked={repeatRestAfterLastDefault}
                  onChange={(e) =>
                    onRepeatRestAfterLastDefaultChange(e.target.checked)
                  }
                />
                <span>Repeat rest after last (default)</span>
              </label>
            </div>
          )}
          {profileTab === "password" && !authHeaderEnabled && (
            <PasswordForm onSubmit={onPasswordChange} />
          )}
          {profileTab === "transfer" && (
            <div className="stack">
              <div className="label">Workout transfer</div>
              <div className="field">
                <label>Export workout</label>
                <select
                  value={exportWorkoutId}
                  onChange={(e) => onExportWorkoutChange(e.target.value)}
                >
                  <option value="">Select workout</option>
                  {activeWorkouts.map((w) => (
                    <option key={w.id} value={w.id}>
                      {w.name}
                    </option>
                  ))}
                </select>
              </div>
              <div className="btn-group">
                <button
                  className="btn subtle"
                  type="button"
                  onClick={onExportWorkout}
                  disabled={!canExport}
                >
                  Export
                </button>
                <button
                  className="btn subtle"
                  type="button"
                  onClick={() => importInputRef.current?.click()}
                >
                  Import
                </button>
              </div>
            </div>
          )}
        </div>
        <div className="profile-tabs">
          <button
            className={profileTab === "settings" ? "tab active" : "tab"}
            onClick={() => onProfileTabChange("settings")}
          >
            Settings
          </button>
          {!authHeaderEnabled && (
            <button
              className={profileTab === "password" ? "tab active" : "tab"}
              onClick={() => onProfileTabChange("password")}
            >
              Password
            </button>
          )}
          <button
            className={profileTab === "transfer" ? "tab active" : "tab"}
            onClick={() => onProfileTabChange("transfer")}
          >
            Export/Import
          </button>
        </div>
      </div>
      {/* Hidden file input for import */}
      <input
        ref={importInputRef}
        type="file"
        accept="application/json"
        style={{ display: "none" }}
        onChange={(e) => {
          const file = e.target.files?.[0];
          if (!file) return;
          onImportWorkout(file);
          e.currentTarget.value = "";
        }}
      />
    </section>
  );
}

// PasswordForm updates the current user's password.
function PasswordForm({
  onSubmit,
}: {
  onSubmit: (currentPassword: string, newPassword: string) => void;
}) {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const mismatch = newPassword !== confirm;
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        // Guard: require matching passwords before submit.
        if (!currentPassword.trim() || !newPassword.trim() || mismatch) return;
        onSubmit(currentPassword.trim(), newPassword.trim());
        setCurrentPassword("");
        setNewPassword("");
        setConfirm("");
      }}
      className="stack"
    >
      <div className="field">
        <label>Current password</label>
        <input
          type="password"
          value={currentPassword}
          onChange={(e) => setCurrentPassword(e.target.value)}
          placeholder="Current password"
          required
        />
      </div>
      <div className="field">
        <label>New password</label>
        <input
          type="password"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
          placeholder="New password"
          required
        />
      </div>
      <div className="field">
        <label>Confirm password</label>
        <input
          type="password"
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          placeholder="Confirm password"
          required
        />
      </div>
      {mismatch && <p className="muted small">Passwords do not match.</p>}
      <button
        className="btn primary"
        type="submit"
        disabled={!currentPassword.trim() || !newPassword.trim() || mismatch}
      >
        Update password
      </button>
    </form>
  );
}
