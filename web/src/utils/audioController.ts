export type AudioController = {
  tokenRef: { current: number };
  unlock: () => Promise<void>;
  pause: () => void;
  resume: () => void;
  stopAll: () => void;
  play: (soundUrl: string) => void;
};

const SILENT_WAV_DATA_URL =
  "data:audio/wav;base64,UklGRiQAAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQAAAAA=";

function createManagedAudio(soundUrl = ""): HTMLAudioElement {
  const audio = new Audio(soundUrl);
  audio.preload = "auto";
  audio.playsInline = true;
  return audio;
}

// createAudioController centralizes play/pause/stop for training audio.
export function createAudioController(): AudioController {
  const activeAudios = new Set<HTMLAudioElement>();
  const pausedAudios = new Set<HTMLAudioElement>();
  const tokenRef = { current: 0 };
  let unlockPromise: Promise<void> | null = null;
  let unlocked = false;

  const stopAudio = (audio: HTMLAudioElement) => {
    audio.pause();
    try {
      audio.currentTime = 0;
    } catch {
      // ignore
    }
    audio.src = "";
    audio.load();
  };

  const unlock = () => {
    if (unlocked || typeof Audio === "undefined") return Promise.resolve();
    if (unlockPromise) return unlockPromise;

    const audio = createManagedAudio(SILENT_WAV_DATA_URL);
    audio.muted = true;

    const cleanup = () => {
      audio.muted = false;
      stopAudio(audio);
      unlockPromise = null;
    };

    unlockPromise = audio
      .play()
      .then(() => {
        unlocked = true;
      })
      .catch(() => {
        // Mobile browsers may still reject if no gesture was active.
      })
      .finally(cleanup);

    return unlockPromise;
  };

  const pause = () => {
    tokenRef.current += 1;

    if (!activeAudios.size) return;
    activeAudios.forEach((audio) => {
      audio.pause();
      pausedAudios.add(audio);
    });
    activeAudios.clear();
  };

  const resume = () => {
    if (!pausedAudios.size) return;
    pausedAudios.forEach((audio) => {
      activeAudios.add(audio);
      audio.play().catch(() => {
        activeAudios.delete(audio);
      });
    });
    pausedAudios.clear();
  };

  const stopAll = () => {
    tokenRef.current += 1;

    if (!activeAudios.size && !pausedAudios.size) return;

    activeAudios.forEach((audio) => stopAudio(audio));
    activeAudios.clear();

    pausedAudios.forEach((audio) => stopAudio(audio));
    pausedAudios.clear();
  };

  const play = (soundUrl: string) => {
    if (!soundUrl || typeof Audio === "undefined") return;

    const audio = createManagedAudio(soundUrl);
    const playToken = tokenRef.current;

    activeAudios.add(audio);
    audio.addEventListener(
      "ended",
      () => {
        activeAudios.delete(audio);
        pausedAudios.delete(audio);
      },
      { once: true },
    );

    const startPlayback = () => {
      if (tokenRef.current !== playToken) {
        activeAudios.delete(audio);
        pausedAudios.delete(audio);
        return;
      }

      audio.play().catch((error) => {
        activeAudios.delete(audio);
        pausedAudios.delete(audio);
        console.warn("audio playback failed", error);
      });
    };

    if (unlockPromise) {
      void unlockPromise.finally(startPlayback);
      return;
    }

    startPlayback();
  };

  return { tokenRef, unlock, pause, resume, stopAll, play };
}
