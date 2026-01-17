import { useEffect, useState } from "react";

type UseUserDefaultsArgs = {
  currentUserId: string | null;
};

// useUserDefaults manages per-user default preferences stored in localStorage.
export function useUserDefaults({ currentUserId }: UseUserDefaultsArgs) {
  const [defaultStepSoundKey, setDefaultStepSoundKey] = useState("");
  const [defaultPauseDuration, setDefaultPauseDuration] = useState("");
  const [defaultPauseSoundKey, setDefaultPauseSoundKey] = useState("");
  const [defaultPauseAutoAdvance, setDefaultPauseAutoAdvance] = useState(false);
  const [repeatRestAfterLastDefault, setRepeatRestAfterLastDefault] =
    useState(false);
  const [pauseOnTabHidden, setPauseOnTabHidden] = useState(false);

  useEffect(() => {
    if (!currentUserId) {
      setRepeatRestAfterLastDefault(false);
      setDefaultStepSoundKey("");
      setDefaultPauseDuration("");
      setDefaultPauseSoundKey("");
      setDefaultPauseAutoAdvance(false);
      setPauseOnTabHidden(false);
      return;
    }

    setRepeatRestAfterLastDefault(
      localStorage.getItem(`motus:repeatRestAfterLast:${currentUserId}`) ===
        "true",
    );
    setDefaultStepSoundKey(
      localStorage.getItem(`motus:defaultStepSound:${currentUserId}`) || "",
    );
    setDefaultPauseDuration(
      localStorage.getItem(`motus:defaultPauseDuration:${currentUserId}`) || "",
    );
    setDefaultPauseSoundKey(
      localStorage.getItem(`motus:defaultPauseSound:${currentUserId}`) || "",
    );
    setDefaultPauseAutoAdvance(
      localStorage.getItem(`motus:defaultPauseAuto:${currentUserId}`) ===
        "true",
    );
    setPauseOnTabHidden(
      localStorage.getItem(`motus:pauseOnHidden:${currentUserId}`) === "true",
    );
  }, [currentUserId]);

  const updateRepeatRestAfterLastDefault = (value: boolean) => {
    setRepeatRestAfterLastDefault(value);
    if (!currentUserId) return;
    localStorage.setItem(
      `motus:repeatRestAfterLast:${currentUserId}`,
      value ? "true" : "false",
    );
  };

  const updateDefaultStepSoundKey = (value: string) => {
    setDefaultStepSoundKey(value);
    if (!currentUserId) return;
    localStorage.setItem(`motus:defaultStepSound:${currentUserId}`, value);
  };

  const updateDefaultPauseDuration = (value: string) => {
    setDefaultPauseDuration(value);
    if (!currentUserId) return;
    localStorage.setItem(`motus:defaultPauseDuration:${currentUserId}`, value);
  };

  const updateDefaultPauseSoundKey = (value: string) => {
    setDefaultPauseSoundKey(value);
    if (!currentUserId) return;
    localStorage.setItem(`motus:defaultPauseSound:${currentUserId}`, value);
  };

  const updateDefaultPauseAutoAdvance = (value: boolean) => {
    setDefaultPauseAutoAdvance(value);
    if (!currentUserId) return;
    localStorage.setItem(
      `motus:defaultPauseAuto:${currentUserId}`,
      value ? "true" : "false",
    );
  };

  const updatePauseOnTabHidden = (value: boolean) => {
    setPauseOnTabHidden(value);
    if (!currentUserId) return;
    localStorage.setItem(
      `motus:pauseOnHidden:${currentUserId}`,
      value ? "true" : "false",
    );
  };

  return {
    defaultStepSoundKey,
    defaultPauseDuration,
    defaultPauseSoundKey,
    defaultPauseAutoAdvance,
    repeatRestAfterLastDefault,
    pauseOnTabHidden,
    updateRepeatRestAfterLastDefault,
    updateDefaultStepSoundKey,
    updateDefaultPauseDuration,
    updateDefaultPauseSoundKey,
    updateDefaultPauseAutoAdvance,
    updatePauseOnTabHidden,
  };
}
