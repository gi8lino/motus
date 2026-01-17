import { useCallback, useEffect, useRef } from "react";
import type { MutableRefObject } from "react";

import type { SoundOption, TrainingState, TrainingStepState } from "../types";
import { resolveMediaUrl } from "../utils/basePath";
import { parseDurationSeconds } from "../utils/time";
import { PROMPTS } from "../utils/messages";
import { createAudioController } from "../utils/audioController";

type TrainingRefs = {
  trainingRef: MutableRefObject<TrainingState | null>;
  currentStepRef: MutableRefObject<TrainingStepState | null>;
  elapsedRef: MutableRefObject<number>;
};

type SoundSchedule = {
  key: string | null;
  triggerAtMs: number;
  soundUrl: string;
};

type UseTrainingAudioArgs = {
  training: TrainingState | null;
  currentStep: TrainingStepState | null;
  sounds: SoundOption[];
  markSoundPlayed: () => void;
  onPause: () => void;
  onStartStep: () => void;
  onToast: (message: string) => void;
  pauseOnTabHidden: boolean;
  refs: TrainingRefs;
};

// useTrainingAudio manages audio playback, scheduling, and tab-hidden pausing.
export function useTrainingAudio({
  training,
  currentStep,
  sounds,
  markSoundPlayed,
  onPause,
  onStartStep,
  onToast,
  pauseOnTabHidden,
  refs,
}: UseTrainingAudioArgs) {
  const { trainingRef, currentStepRef, elapsedRef } = refs;

  const audioRef = useRef(createAudioController());
  const {
    tokenRef,
    pause: pauseAudio,
    resume: resumeAudio,
    stopAll: stopAllAudio,
    play: playAudio,
  } = audioRef.current;

  const stepSoundTimerRef = useRef<number | null>(null);
  const subsetSoundTimerRef = useRef<number | null>(null);

  const stepSoundScheduleRef = useRef<SoundSchedule>({
    key: null,
    triggerAtMs: 0,
    soundUrl: "",
  });
  const subsetSoundScheduleRef = useRef<SoundSchedule>({
    key: null,
    triggerAtMs: 0,
    soundUrl: "",
  });

  const trainingIdRef = useRef<string | null>(null);
  const subsetSoundPlayedRef = useRef(new Set<string>());

  const clearSoundTimers = useCallback(() => {
    if (stepSoundTimerRef.current) {
      clearTimeout(stepSoundTimerRef.current);
      stepSoundTimerRef.current = null;
    }
    if (subsetSoundTimerRef.current) {
      clearTimeout(subsetSoundTimerRef.current);
      subsetSoundTimerRef.current = null;
    }
  }, []);

  const handlePause = useCallback(() => {
    pauseAudio();
    onPause();
  }, [onPause, pauseAudio]);

  const handleStart = useCallback(() => {
    onStartStep();
    resumeAudio();
  }, [onStartStep, resumeAudio]);

  // Pause active training when the tab is hidden (configurable).
  const hiddenPauseNotifiedRef = useRef(false);

  useEffect(() => {
    if (!pauseOnTabHidden) {
      hiddenPauseNotifiedRef.current = false;
      return;
    }

    const handleVisibility = () => {
      const s = trainingRef.current;
      if (document.hidden) {
        if (s?.running) {
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
  }, [pauseOnTabHidden, onPause, onToast, pauseAudio, trainingRef]);

  // Reset subset-level tracking on training changes.
  useEffect(() => {
    const currentTrainingId = training?.trainingId || null;
    if (trainingIdRef.current !== currentTrainingId) {
      trainingIdRef.current = currentTrainingId;
      subsetSoundPlayedRef.current = new Set();
    }
  }, [training?.trainingId]);

  // Cleanup audio and timers on unmount.
  useEffect(() => {
    return () => {
      clearSoundTimers();
      stopAllAudio();
      stepSoundScheduleRef.current = {
        key: null,
        triggerAtMs: 0,
        soundUrl: "",
      };
      subsetSoundScheduleRef.current = {
        key: null,
        triggerAtMs: 0,
        soundUrl: "",
      };
    };
  }, [clearSoundTimers, stopAllAudio]);

  // Schedule subset + exercise target sounds.
  useEffect(() => {
    if (!currentStep || !training?.running) {
      clearSoundTimers();
      if (!currentStep) stopAllAudio();
      else pauseAudio();
      stepSoundScheduleRef.current = {
        key: null,
        triggerAtMs: 0,
        soundUrl: "",
      };
      subsetSoundScheduleRef.current = {
        key: null,
        triggerAtMs: 0,
        soundUrl: "",
      };
      return;
    }

    const subsetSoundKey = currentStep.soundKey || "";
    const exerciseSoundKey =
      currentStep.exercises?.length === 1
        ? currentStep.exercises[0]?.soundKey || ""
        : "";

    const subsetSoundOpt = subsetSoundKey
      ? sounds.find((sound) => sound.key === subsetSoundKey)
      : sounds.find((sound) => sound.file === currentStep.soundUrl);

    const subsetSoundUrl = resolveMediaUrl(
      currentStep.soundUrl || subsetSoundOpt?.file || "",
    );

    const exerciseSoundOpt = exerciseSoundKey
      ? sounds.find((sound) => sound.key === exerciseSoundKey)
      : undefined;

    const exerciseSoundUrl =
      exerciseSoundKey && exerciseSoundOpt
        ? resolveMediaUrl(exerciseSoundOpt.file || "")
        : subsetSoundUrl;

    if (!subsetSoundUrl && !exerciseSoundUrl) return;

    const subsetLeadSeconds = subsetSoundOpt?.leadSeconds ?? 0;
    const exerciseLeadSeconds = exerciseSoundKey
      ? (exerciseSoundOpt?.leadSeconds ?? 0)
      : subsetLeadSeconds;

    const subsetTargetSeconds = currentStep.subsetEstimatedSeconds ?? 0;
    const hasSubsetTarget = subsetTargetSeconds > 0;

    const exerciseTargetSeconds =
      currentStep.estimatedSeconds ||
      parseDurationSeconds((currentStep as any).duration) ||
      parseDurationSeconds(currentStep.exercises?.[0]?.duration) ||
      0;

    const allowStepSound = !currentStep.soundPlayed;

    const tokenAtSchedule = tokenRef.current;

    // Subset target sound: schedule once per subset instance.
    if (hasSubsetTarget && currentStep.subsetId && subsetSoundUrl) {
      const subsetInstanceId = `${currentStep.subsetId}-${currentStep.loopIndex ?? 0}`;

      if (!subsetSoundPlayedRef.current.has(subsetInstanceId)) {
        const subsetTargetMs = subsetTargetSeconds * 1000;
        const leadMs = Math.min(
          subsetTargetMs,
          Math.max(0, subsetLeadSeconds * 1000),
        );
        const triggerMs = Math.max(0, subsetTargetMs - leadMs);

        const scheduleKey = `${subsetInstanceId}:${subsetTargetMs}:${subsetSoundUrl}`;
        const existing = subsetSoundScheduleRef.current;

        if (existing.key !== scheduleKey) {
          if (subsetSoundTimerRef.current) {
            clearTimeout(subsetSoundTimerRef.current);
            subsetSoundTimerRef.current = null;
          }

          subsetSoundScheduleRef.current = {
            key: scheduleKey,
            triggerAtMs: 0,
            soundUrl: subsetSoundUrl,
          };

          const subsetElapsedMs = (() => {
            const s = trainingRef.current;
            const stepNow = currentStepRef.current;
            if (!s || !stepNow?.subsetId) return 0;

            const currentElapsed = elapsedRef.current;
            const targetLoop = stepNow.loopIndex ?? 0;

            return s.steps.reduce((acc, step, idx) => {
              if (step.subsetId !== stepNow.subsetId) return acc;
              const stepLoop = step.loopIndex ?? 0;
              if (stepLoop !== targetLoop) return acc;
              if (idx < s.currentIndex) return acc + (step.elapsedMillis || 0);
              if (idx === s.currentIndex) return acc + currentElapsed;
              return acc;
            }, 0);
          })();

          const remaining = triggerMs - subsetElapsedMs;

          const fire = () => {
            if (tokenRef.current !== tokenAtSchedule) return;

            stopAllAudio();
            playAudio(subsetSoundUrl);
            subsetSoundTimerRef.current = null;
            subsetSoundPlayedRef.current.add(subsetInstanceId);
          };

          if (remaining <= 0) fire();
          else {
            subsetSoundTimerRef.current = window.setTimeout(fire, remaining);
            subsetSoundScheduleRef.current.triggerAtMs = Date.now() + remaining;
          }
        }
      }
    }

    // Step/exercise target sound: schedule once per step instance + config.
    if (allowStepSound && exerciseTargetSeconds > 0 && exerciseSoundUrl) {
      const stepTargetMs = exerciseTargetSeconds * 1000;
      const leadMs = Math.min(
        stepTargetMs,
        Math.max(0, exerciseLeadSeconds * 1000),
      );
      const triggerMs = Math.max(0, stepTargetMs - leadMs);

      const scheduleKey = `${currentStep.id || `${training.trainingId}-${training.currentIndex}`}:${stepTargetMs}:${exerciseSoundUrl}`;
      const existing = stepSoundScheduleRef.current;

      if (existing.key !== scheduleKey) {
        if (stepSoundTimerRef.current) {
          clearTimeout(stepSoundTimerRef.current);
          stepSoundTimerRef.current = null;
        }

        stepSoundScheduleRef.current = {
          key: scheduleKey,
          triggerAtMs: 0,
          soundUrl: exerciseSoundUrl,
        };

        const remaining = triggerMs - elapsedRef.current;

        const fire = () => {
          if (tokenRef.current !== tokenAtSchedule) return;

          stopAllAudio();
          playAudio(exerciseSoundUrl);
          stepSoundTimerRef.current = null;
          markSoundPlayed();
        };

        if (remaining <= 0) fire();
        else {
          stepSoundTimerRef.current = window.setTimeout(fire, remaining);
          stepSoundScheduleRef.current.triggerAtMs = Date.now() + remaining;
        }
      }
    }
  }, [
    currentStep?.id,
    currentStep?.soundKey,
    currentStep?.soundUrl,
    currentStep?.estimatedSeconds,
    currentStep?.subsetEstimatedSeconds,
    (currentStep as any)?.duration,
    currentStep?.exercises?.[0]?.duration,
    currentStep?.exercises?.[0]?.soundKey,
    currentStep?.subsetId,
    currentStep?.loopIndex,
    training?.trainingId,
    training?.currentIndex,
    training?.running,
    training?.steps,
    sounds,
    markSoundPlayed,
    pauseAudio,
    stopAllAudio,
    playAudio,
    clearSoundTimers,
    elapsedRef,
    trainingRef,
    currentStepRef,
  ]);

  return {
    handlePause,
    handleStart,
    pauseActiveAudio: pauseAudio,
    resumePausedAudio: resumeAudio,
    stopActiveAudio: stopAllAudio,
  };
}
