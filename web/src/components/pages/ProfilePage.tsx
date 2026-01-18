import { useEffect, useState } from "react";
import type { RefObject } from "react";
import type { SoundOption, Workout } from "../../types";
import { SelectDropdown } from "../common/SelectDropdown";
import { MESSAGES, toErrorMessage } from "../../utils/messages";
import { UI_TEXT } from "../../utils/uiText";

type ProfileTab = "settings" | "password" | "transfer";
type ThemeMode = "auto" | "dark" | "light";

export type ProfileViewData = {
  profileTab: ProfileTab;
  currentName: string;
  themeMode: ThemeMode;
  sounds: SoundOption[];
  defaultStepSoundKey: string;
  defaultPauseDuration: string;
  defaultPauseSoundKey: string;
  defaultPauseAutoAdvance: boolean;
  repeatRestAfterLastDefault: boolean;
  pauseOnTabHidden: boolean;
  showHours: boolean;
  exportWorkoutId: string;
  activeWorkouts: Workout[];
  importInputRef: RefObject<HTMLInputElement | null>;
  authHeaderEnabled: boolean;
};

export type ProfileViewActions = {
  onProfileTabChange: (tab: ProfileTab) => void;
  onUpdateName: (name: string) => void | Promise<void>;
  onThemeChange: (mode: ThemeMode) => void;
  onDefaultStepSoundChange: (value: string) => void;
  onDefaultPauseDurationChange: (value: string) => void;
  onDefaultPauseSoundChange: (value: string) => void;
  onDefaultPauseAutoAdvanceChange: (value: boolean) => void;
  onRepeatRestAfterLastDefaultChange: (value: boolean) => void;
  onPauseOnTabHiddenChange: (value: boolean) => void;
  onShowHoursChange: (value: boolean) => void;
  onExportWorkoutChange: (id: string) => void;
  onExportWorkout: () => void | Promise<void>;
  onImportWorkout: (file: File) => void | Promise<void>;
  onPasswordChange: (
    currentPassword: string,
    newPassword: string,
  ) => void | Promise<void>;
};

// ProfileView renders account preferences and transfer actions.
export function ProfileView({
  data,
  actions,
}: {
  data: ProfileViewData;
  actions: ProfileViewActions;
}) {
  const {
    profileTab,
    currentName,
    themeMode,
    sounds,
    defaultStepSoundKey,
    defaultPauseDuration,
    defaultPauseSoundKey,
    defaultPauseAutoAdvance,
    repeatRestAfterLastDefault,
    pauseOnTabHidden,
    showHours,
    exportWorkoutId,
    activeWorkouts,
    importInputRef,
    authHeaderEnabled,
  } = data;
  const {
    onProfileTabChange,
    onUpdateName,
    onThemeChange,
    onDefaultStepSoundChange,
    onDefaultPauseDurationChange,
    onDefaultPauseSoundChange,
    onDefaultPauseAutoAdvanceChange,
    onRepeatRestAfterLastDefaultChange,
    onPauseOnTabHiddenChange,
    onShowHoursChange,
    onExportWorkoutChange,
    onExportWorkout,
    onImportWorkout,
    onPasswordChange,
  } = actions;
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
          <h3>{UI_TEXT.pages.profile.title}</h3>
          <p className="muted small hint">{UI_TEXT.pages.profile.hint}</p>
        </div>
      </div>
      <div className="profile-layout">
        <div className="profile-content">
          {/* Tab content */}
          {profileTab === "settings" && (
            <div className="stack">
              <div className="label">{UI_TEXT.pages.profile.accountLabel}</div>
              <DisplayNameForm
                currentName={currentName}
                onUpdate={onUpdateName}
              />
              <div className="divider" />
              <div className="label">
                {UI_TEXT.pages.profile.appearanceLabel}
              </div>
              <div className="field">
                <label>{UI_TEXT.pages.profile.themeLabel}</label>
                <SelectDropdown
                  items={[
                    { id: "auto", label: UI_TEXT.pages.profile.autoTheme },
                    { id: "dark", label: UI_TEXT.pages.profile.darkTheme },
                    { id: "light", label: UI_TEXT.pages.profile.lightTheme },
                  ]}
                  value={themeMode}
                  placeholder={UI_TEXT.placeholders.selectTheme}
                  onSelect={(item) => onThemeChange(item.id as ThemeMode)}
                />
              </div>
              <div className="divider" />
              <div className="label">{UI_TEXT.pages.profile.defaultsLabel}</div>
              <div className="field">
                <label>{UI_TEXT.pages.profile.defaultStepSoundLabel}</label>
                <SelectDropdown
                  items={[
                    { id: "", label: UI_TEXT.options.none },
                    ...sounds.map((sound) => ({
                      id: sound.key,
                      label: sound.label,
                    })),
                  ]}
                  value={defaultStepSoundKey || null}
                  placeholder={UI_TEXT.placeholders.selectSound}
                  onSelect={(item) => onDefaultStepSoundChange(item.id)}
                  onClear={
                    defaultStepSoundKey
                      ? () => onDefaultStepSoundChange("")
                      : undefined
                  }
                />
              </div>
              <div className="field">
                <label>{UI_TEXT.pages.profile.defaultPauseDurationLabel}</label>
                <input
                  value={defaultPauseDuration}
                  onChange={(e) => onDefaultPauseDurationChange(e.target.value)}
                  placeholder={UI_TEXT.placeholders.duration}
                />
              </div>
              <div className="field">
                <label>{UI_TEXT.pages.profile.defaultPauseSoundLabel}</label>
                <SelectDropdown
                  items={[
                    { id: "", label: UI_TEXT.options.none },
                    ...sounds.map((sound) => ({
                      id: sound.key,
                      label: sound.label,
                    })),
                  ]}
                  value={defaultPauseSoundKey || null}
                  placeholder={UI_TEXT.placeholders.selectSound}
                  onSelect={(item) => onDefaultPauseSoundChange(item.id)}
                  onClear={
                    defaultPauseSoundKey
                      ? () => onDefaultPauseSoundChange("")
                      : undefined
                  }
                />
              </div>
              <label
                className="switch"
                title={UI_TEXT.titles.defaultPauseAutoAdvance}
              >
                <input
                  type="checkbox"
                  checked={defaultPauseAutoAdvance}
                  onChange={(e) =>
                    onDefaultPauseAutoAdvanceChange(e.target.checked)
                  }
                />
                <span className="switch-slider" aria-hidden="true" />
                <span className="switch-label">
                  {UI_TEXT.pages.profile.defaultPauseAutoAdvanceLabel}
                </span>
              </label>
              <label
                className="switch"
                title={UI_TEXT.titles.repeatRestAfterLast}
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
                  {UI_TEXT.pages.profile.repeatRestAfterLastLabel}
                </span>
              </label>
              <label className="switch" title={UI_TEXT.titles.pauseOnTabHidden}>
                <input
                  type="checkbox"
                  checked={pauseOnTabHidden}
                  onChange={(e) => onPauseOnTabHiddenChange(e.target.checked)}
                />
                <span className="switch-slider" aria-hidden="true" />
                <span className="switch-label">
                  {UI_TEXT.pages.profile.pauseOnTabHiddenLabel}
                </span>
              </label>
              <label className="switch" title={UI_TEXT.titles.showHours}>
                <input
                  type="checkbox"
                  checked={showHours}
                  onChange={(e) => onShowHoursChange(e.target.checked)}
                />
                <span className="switch-slider" aria-hidden="true" />
                <span className="switch-label">
                  {UI_TEXT.pages.profile.showHoursLabel}
                </span>
              </label>
            </div>
          )}
          {profileTab === "password" && !authHeaderEnabled && (
            <PasswordForm onSubmit={onPasswordChange} />
          )}
          {profileTab === "transfer" && (
            <div className="stack">
              <div className="label">{UI_TEXT.pages.profile.transferLabel}</div>
              <div className="field">
                <label>{UI_TEXT.pages.profile.exportLabel}</label>
                <SelectDropdown
                  items={activeWorkouts.map((workout) => ({
                    id: workout.id,
                    label: workout.name,
                  }))}
                  value={exportWorkoutId || null}
                  placeholder={UI_TEXT.placeholders.selectWorkout}
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
                  {UI_TEXT.pages.profile.exportButton}
                </button>
                <button
                  className="btn primary"
                  type="button"
                  onClick={() => importInputRef.current?.click()}
                >
                  {UI_TEXT.pages.profile.importButton}
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
            {UI_TEXT.pages.profile.settingsTab}
          </button>
          {!authHeaderEnabled && (
            <button
              className={profileTab === "password" ? "tab active" : "tab"}
              onClick={() => onProfileTabChange("password")}
            >
              {UI_TEXT.pages.profile.passwordTab}
            </button>
          )}
          <button
            className={profileTab === "transfer" ? "tab active" : "tab"}
            onClick={() => onProfileTabChange("transfer")}
          >
            {UI_TEXT.pages.profile.transferTab}
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
        <label>{UI_TEXT.pages.profile.currentPasswordLabel}</label>
        <input
          type="password"
          value={currentPassword}
          onChange={(e) => setCurrentPassword(e.target.value)}
          placeholder={UI_TEXT.placeholders.currentPassword}
          required
        />
      </div>
      <div className="field">
        <label>{UI_TEXT.pages.profile.newPasswordLabel}</label>
        <input
          type="password"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
          placeholder={UI_TEXT.placeholders.newPassword}
          required
        />
      </div>
      <div className="field">
        <label>{UI_TEXT.pages.profile.confirmPasswordLabel}</label>
        <input
          type="password"
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          placeholder={UI_TEXT.placeholders.confirmPassword}
          required
        />
      </div>
      {mismatch && (
        <p className="muted small">{UI_TEXT.pages.profile.mismatch}</p>
      )}
      <button
        className="btn primary"
        type="submit"
        disabled={!currentPassword.trim() || !newPassword.trim() || mismatch}
      >
        {UI_TEXT.pages.profile.updatePasswordButton}
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
      <label>{UI_TEXT.pages.profile.displayNameLabel}</label>
      <div className="input-row">
        <input
          value={name}
          onChange={(e) => {
            setName(e.target.value);
            setError(null);
          }}
          placeholder={UI_TEXT.placeholders.yourName}
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
          {saving ? UI_TEXT.account.saving : UI_TEXT.account.save}
        </button>
      </div>
      {error && <div className="muted small">{error}</div>}
    </div>
  );
}
