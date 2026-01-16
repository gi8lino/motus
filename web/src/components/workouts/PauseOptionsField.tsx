import type { SoundOption } from "../../types";

type PauseOptionsFieldProps = {
  autoAdvance: boolean;
  soundKey: string;
  sounds: SoundOption[];
  onAutoAdvanceChange: (value: boolean) => void;
  onSoundChange: (soundKey: string) => void;
  extra?: React.ReactNode;
};

// PauseOptionsField renders pause-specific options (auto-advance + sound).
export function PauseOptionsField({
  autoAdvance,
  soundKey,
  sounds,
  onAutoAdvanceChange,
  onSoundChange,
  extra,
}: PauseOptionsFieldProps) {
  return (
    <div className="stack" style={{ gap: 10 }}>
      <label
        className="switch"
        title="Auto-advance ends the pause automatically."
      >
        <input
          type="checkbox"
          checked={Boolean(autoAdvance)}
          onChange={(e) => onAutoAdvanceChange(e.target.checked)}
        />
        <span className="switch-slider" aria-hidden="true" />
        <span className="switch-label">Auto-advance</span>
      </label>

      <div className="field">
        <label>Sound</label>
        <select
          value={soundKey || ""}
          onChange={(e) => onSoundChange(e.target.value)}
        >
          <option value="">None</option>
          {sounds.map((sound) => (
            <option key={sound.key} value={sound.key}>
              {sound.label}
            </option>
          ))}
        </select>
      </div>

      {extra ?? null}
    </div>
  );
}
