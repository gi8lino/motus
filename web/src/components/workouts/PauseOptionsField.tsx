import type { SoundOption } from "../../types";

// PauseOptionsField renders shared pause controls (auto-advance + sound).
export function PauseOptionsField({
  autoAdvance,
  soundKey,
  sounds,
  onAutoAdvanceChange,
  onSoundChange,
  extra,
}: {
  autoAdvance: boolean;
  soundKey: string;
  sounds: SoundOption[];
  onAutoAdvanceChange: (value: boolean) => void;
  onSoundChange: (value: string) => void;
  extra?: React.ReactNode;
}) {
  return (
    <>
      <label className="field checkbox">
        <input
          type="checkbox"
          checked={autoAdvance}
          onChange={(e) => onAutoAdvanceChange(e.target.checked)}
        />
        <span>Auto-advance when time elapses</span>
      </label>
      <div className="field">
        <label>Sound</label>
        <select
          value={soundKey}
          onChange={(e) => onSoundChange(e.target.value)}
        >
          <option value="">None</option>
          {sounds.map((s) => (
            <option key={s.key} value={s.key}>
              {s.label}
            </option>
          ))}
        </select>
      </div>
      {extra}
    </>
  );
}
