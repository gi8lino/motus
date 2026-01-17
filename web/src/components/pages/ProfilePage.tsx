import { useEffect, useState } from "react";
import type { RefObject } from "react";
import type { SoundOption, Workout } from "../../types";
import { SelectDropdown } from "../common/SelectDropdown";
import { MESSAGES, toErrorMessage } from "../../utils/messages";

type ProfileTab = "settings" | "password" | "transfer";
type ThemeMode = "auto" | "dark" | "light";

// ProfileView renders account preferences and transfer actions.
export function ProfileView({
  profileTab,
  onProfileTabChange,
  currentName,
  onUpdateName,
  themeMode,
  onThemeChange,
  sounds,
  defaultStepSoundKey,
  onDefaultStepSoundChange,
  defaultPauseDuration,
  onDefaultPauseDurationChange,
  defaultPauseSoundKey,
  onDefaultPauseSoundChange,
  defaultPauseAutoAdvance,
  onDefaultPauseAutoAdvanceChange,
  repeatRestAfterLastDefault,
  onRepeatRestAfterLastDefaultChange,
  pauseOnTabHidden,
  onPauseOnTabHiddenChange,
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
  currentName: string;
  onUpdateName: (name: string) => void | Promise<void>;
  themeMode: ThemeMode;
  onThemeChange: (mode: ThemeMode) => void;
  sounds: SoundOption[];
  defaultStepSoundKey: string;
  onDefaultStepSoundChange: (value: string) => void;
  defaultPauseDuration: string;
  onDefaultPauseDurationChange: (value: string) => void;
  defaultPauseSoundKey: string;
  onDefaultPauseSoundChange: (value: string) => void;
  defaultPauseAutoAdvance: boolean;
  onDefaultPauseAutoAdvanceChange: (value: boolean) => void;
  repeatRestAfterLastDefault: boolean;
  onRepeatRestAfterLastDefaultChange: (value: boolean) => void;
  pauseOnTabHidden: boolean;
  onPauseOnTabHiddenChange: (value: boolean) => void;
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
  // Prevent password tab access when auth headers are enabled.
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
          <p className="muted small hint">
            Manage your local account preferences.
          </p>
        </div>
      </div>
      <div className="profile-layout">
        <div className="profile-content">
          {/* Tab content */}
          {profileTab === "settings" && (
            <div className="stack">
              <div className="label">Account</div>
              <DisplayNameForm
                currentName={currentName}
                onUpdate={onUpdateName}
              />
              <div className="divider" />
              <div className="label">Appearance</div>
              <div className="field">
                <label>Theme</label>
                <SelectDropdown
                  items={[
                    { id: "auto", label: "Auto (system)" },
                    { id: "dark", label: "Dark" },
                    { id: "light", label: "Light" },
                  ]}
                  value={themeMode}
                  placeholder="Select theme"
                  onSelect={(item) => onThemeChange(item.id as ThemeMode)}
                />
              </div>
              <div className="divider" />
              <div className="label">Defaults</div>
              <div className="field">
                <label>Default step sound</label>
                <SelectDropdown
                  items={[
                    { id: "", label: "None" },
                    ...sounds.map((sound) => ({
                      id: sound.key,
                      label: sound.label,
                    })),
                  ]}
                  value={defaultStepSoundKey || null}
                  placeholder="Select sound"
                  onSelect={(item) => onDefaultStepSoundChange(item.id)}
                  onClear={
                    defaultStepSoundKey
                      ? () => onDefaultStepSoundChange("")
                      : undefined
                  }
                />
              </div>
              <div className="field">
                <label>Default pause duration</label>
                <input
                  value={defaultPauseDuration}
                  onChange={(e) => onDefaultPauseDurationChange(e.target.value)}
                  placeholder="e.g. 45s"
                />
              </div>
              <div className="field">
                <label>Default pause sound</label>
                <SelectDropdown
                  items={[
                    { id: "", label: "None" },
                    ...sounds.map((sound) => ({
                      id: sound.key,
                      label: sound.label,
                    })),
                  ]}
                  value={defaultPauseSoundKey || null}
                  placeholder="Select sound"
                  onSelect={(item) => onDefaultPauseSoundChange(item.id)}
                  onClear={
                    defaultPauseSoundKey
                      ? () => onDefaultPauseSoundChange("")
                      : undefined
                  }
                />
              </div>
              <label className="switch" title="Auto-advance pauses by default">
                <input
                  type="checkbox"
                  checked={defaultPauseAutoAdvance}
                  onChange={(e) =>
                    onDefaultPauseAutoAdvanceChange(e.target.checked)
                  }
                />
                <span className="switch-slider" aria-hidden="true" />
                <span className="switch-label">Default pause auto-advance</span>
              </label>
              <label
                className="switch"
                title="Repeat the rest step after the final round"
              >
                <input
                  type="checkbox"
                  checked={repeatRestAfterLastDefault}
                  onChange={(e) =>
                    onRepeatRestAfterLastDefaultChange(e.target.checked)
                  }
                />
                <span className="switch-slider" aria-hidden="true" />
                <span className="switch-label">
                  Repeat rest after last (default)
                </span>
              </label>
              <label
                className="switch"
                title="Pause training when tab is hidden"
              >
                <input
                  type="checkbox"
                  checked={pauseOnTabHidden}
                  onChange={(e) => onPauseOnTabHiddenChange(e.target.checked)}
                />
                <span className="switch-slider" aria-hidden="true" />
                <span className="switch-label">
                  Pause training when tab is hidden
                </span>
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
                <SelectDropdown
                  items={activeWorkouts.map((workout) => ({
                    id: workout.id,
                    label: workout.name,
                  }))}
                  value={exportWorkoutId || null}
                  placeholder="Select workout"
                  onSelect={(item) => onExportWorkoutChange(item.id)}
                  onClear={
                    exportWorkoutId
                      ? () => onExportWorkoutChange("")
                      : undefined
                  }
                />
              </div>
              <div className="btn-group">
                <button
                  className={canExport ? "btn primary" : "btn subtle"}
                  type="button"
                  onClick={onExportWorkout}
                  disabled={!canExport}
                >
                  Export
                </button>
                <button
                  className="btn primary"
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

// DisplayNameForm edits the user-friendly display name.
function DisplayNameForm({
  currentName,
  onUpdate,
}: {
  currentName: string;
  onUpdate: (name: string) => void | Promise<void>;
}) {
  const [name, setName] = useState(currentName);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Keep the display name input in sync with prop changes.
  useEffect(() => {
    setName(currentName);
  }, [currentName]);

  const trimmed = name.trim();
  const canSave = trimmed.length > 0 && trimmed !== currentName;

  return (
    <div className="field">
      <label>Display name</label>
      <div className="input-row">
        <input
          value={name}
          onChange={(e) => {
            setName(e.target.value);
            setError(null);
          }}
          placeholder="Your name"
        />
        <button
          className={canSave ? "btn primary" : "btn subtle"}
          type="button"
          disabled={!canSave || saving}
          onClick={async () => {
            setSaving(true);
            setError(null);
            try {
              await onUpdate(trimmed);
            } catch (err) {
              setError(toErrorMessage(err, MESSAGES.updateNameFailed));
            } finally {
              setSaving(false);
            }
          }}
        >
          {saving ? "Saving…" : "Save"}
        </button>
      </div>
      {error && <div className="muted small">{error}</div>}
    </div>
  );
}
