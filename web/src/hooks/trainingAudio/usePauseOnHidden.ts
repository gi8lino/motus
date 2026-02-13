import { useEffect, useRef } from "react";
import type { MutableRefObject } from "react";
import type { TrainingState } from "../../types";
import { PROMPTS } from "../../utils/messages";

type UsePauseOnHiddenArgs = {
  pauseOnTabHidden: boolean;
  trainingRef: MutableRefObject<TrainingState | null>;
  pauseAudio: () => void;
  onPause: () => void;
  onToast: (message: string) => void;
};

// usePauseOnHidden pauses active training when the tab becomes hidden.
export function usePauseOnHidden({
  pauseOnTabHidden,
  trainingRef,
  pauseAudio,
  onPause,
  onToast,
}: UsePauseOnHiddenArgs): void {
  const hiddenPauseNotifiedRef = useRef(false);

  useEffect(() => {
    if (!pauseOnTabHidden) {
      hiddenPauseNotifiedRef.current = false;
      return;
    }

    const handleVisibility = () => {
      const current = trainingRef.current;
      if (document.hidden) {
        if (current?.running) {
          pauseAudio();
          onPause();
          if (!hiddenPauseNotifiedRef.current) {
            onToast(PROMPTS.trainingPausedHidden);
            hiddenPauseNotifiedRef.current = true;
          }
        }
      } else {
        hiddenPauseNotifiedRef.current = false;
      }
    };

    document.addEventListener("visibilitychange", handleVisibility);
    return () =>
      document.removeEventListener("visibilitychange", handleVisibility);
  }, [pauseOnTabHidden, trainingRef, pauseAudio, onPause, onToast]);
}
