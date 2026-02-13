import { useCallback, useEffect, useRef } from "react";

import { createAudioController } from "../utils/audioController";
import {
  clearTimer,
  getSubsetElapsedMs,
  resolveSoundPlan,
} from "./trainingAudio/helpers";
import type {
  SoundSchedule,
  UseTrainingAudioArgs,
} from "./trainingAudio/types";
import { usePauseOnHidden } from "./trainingAudio/usePauseOnHidden";

// useTrainingAudio manages training audio playback and target-sound scheduling.
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
    clearTimer(stepSoundTimerRef);
    clearTimer(subsetSoundTimerRef);
  }, []);

  const resetSchedules = useCallback(() => {
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
  }, []);

  const handlePause = useCallback(() => {
    pauseAudio();
    onPause();
  }, [onPause, pauseAudio]);

  const handleStart = useCallback(() => {
    onStartStep();
    resumeAudio();
  }, [onStartStep, resumeAudio]);

  usePauseOnHidden({
    pauseOnTabHidden,
    trainingRef,
    pauseAudio,
    onPause,
    onToast,
  });

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
      resetSchedules();
    };
  }, [clearSoundTimers, stopAllAudio, resetSchedules]);

  // Schedule subset + exercise target sounds.
  useEffect(() => {
    if (!currentStep || !training?.running) {
      clearSoundTimers();
      if (!currentStep) stopAllAudio();
      else pauseAudio();
      resetSchedules();
      return;
    }

    const soundPlan = resolveSoundPlan(currentStep, sounds);
    if (!soundPlan.subsetSoundUrl && !soundPlan.exerciseSoundUrl) return;

    const hasSubsetTarget = soundPlan.subsetTargetSeconds > 0;
    const allowStepSound = !currentStep.soundPlayed;
    const tokenAtSchedule = tokenRef.current;

    // Subset target sound: schedule once per subset instance.
    if (hasSubsetTarget && currentStep.subsetId && soundPlan.subsetSoundUrl) {
      const subsetInstanceID = `${currentStep.subsetId}-${currentStep.loopIndex ?? 0}`;

      if (!subsetSoundPlayedRef.current.has(subsetInstanceID)) {
        const subsetTargetMs = soundPlan.subsetTargetSeconds * 1000;
        const leadMs = Math.min(
          subsetTargetMs,
          Math.max(0, soundPlan.subsetLeadSeconds * 1000),
        );
        const triggerMs = Math.max(0, subsetTargetMs - leadMs);

        const scheduleKey = `${subsetInstanceID}:${subsetTargetMs}:${soundPlan.subsetSoundUrl}`;
        const existing = subsetSoundScheduleRef.current;

        if (existing.key !== scheduleKey) {
          clearTimer(subsetSoundTimerRef);
          subsetSoundScheduleRef.current = {
            key: scheduleKey,
            triggerAtMs: 0,
            soundUrl: soundPlan.subsetSoundUrl,
          };

          const trainingState = trainingRef.current;
          const currentStepState = currentStepRef.current;
          const subsetElapsedMs =
            trainingState && currentStepState
              ? getSubsetElapsedMs(
                  trainingState,
                  currentStepState,
                  elapsedRef.current,
                )
              : 0;

          const remaining = triggerMs - subsetElapsedMs;

          const fire = () => {
            if (tokenRef.current !== tokenAtSchedule) return;

            stopAllAudio();
            playAudio(soundPlan.subsetSoundUrl);
            subsetSoundTimerRef.current = null;
            subsetSoundPlayedRef.current.add(subsetInstanceID);
          };

          if (remaining <= 0) {
            fire();
          } else {
            subsetSoundTimerRef.current = window.setTimeout(fire, remaining);
            subsetSoundScheduleRef.current.triggerAtMs = Date.now() + remaining;
          }
        }
      }
    }

    // Step/exercise target sound: schedule once per step instance + config.
    if (
      allowStepSound &&
      soundPlan.exerciseTargetSeconds > 0 &&
      soundPlan.exerciseSoundUrl
    ) {
      const stepTargetMs = soundPlan.exerciseTargetSeconds * 1000;
      const leadMs = Math.min(
        stepTargetMs,
        Math.max(0, soundPlan.exerciseLeadSeconds * 1000),
      );
      const triggerMs = Math.max(0, stepTargetMs - leadMs);

      const scheduleKey = `${currentStep.id || `${training.trainingId}-${training.currentIndex}`}:${stepTargetMs}:${soundPlan.exerciseSoundUrl}`;
      const existing = stepSoundScheduleRef.current;

      if (existing.key !== scheduleKey) {
        clearTimer(stepSoundTimerRef);
        stepSoundScheduleRef.current = {
          key: scheduleKey,
          triggerAtMs: 0,
          soundUrl: soundPlan.exerciseSoundUrl,
        };

        const remaining = triggerMs - elapsedRef.current;

        const fire = () => {
          if (tokenRef.current !== tokenAtSchedule) return;

          stopAllAudio();
          playAudio(soundPlan.exerciseSoundUrl);
          stepSoundTimerRef.current = null;
          markSoundPlayed();
        };

        if (remaining <= 0) {
          fire();
        } else {
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
    currentStep?.duration,
    currentStep?.exercises?.[0]?.duration,
    currentStep?.exercises?.[0]?.soundKey,
    currentStep?.subsetId,
    currentStep?.loopIndex,
    currentStep?.soundPlayed,
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
    resetSchedules,
    elapsedRef,
    trainingRef,
    currentStepRef,
    tokenRef,
  ]);

  return {
    handlePause,
    handleStart,
    pauseActiveAudio: pauseAudio,
    resumePausedAudio: resumeAudio,
    stopActiveAudio: stopAllAudio,
  };
}
