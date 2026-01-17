export type AudioController = {
  tokenRef: { current: number };
  pause: () => void;
  resume: () => void;
  stopAll: () => void;
  play: (soundUrl: string) => void;
};

// createAudioController centralizes play/pause/stop for training audio.
export function createAudioController(): AudioController {
  const activeAudios = new Set<HTMLAudioElement>();
  const pausedAudios = new Set<HTMLAudioElement>();
  const tokenRef = { current: 0 };

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

  const pause = () => {
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
    const audio = new Audio(soundUrl);
    activeAudios.add(audio);
    audio.addEventListener(
      "ended",
      () => {
        activeAudios.delete(audio);
        pausedAudios.delete(audio);
      },
      { once: true },
    );
    audio.play().catch(() => {
      activeAudios.delete(audio);
    });
  };

  return { tokenRef, pause, resume, stopAll, play };
}
