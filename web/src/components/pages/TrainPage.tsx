import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import type {
  TrainngState,
  TrainngStepState,
  SoundOption,
  Workout,
} from "../../types";
import { formatMillis } from "../../utils/format";
import { resolveMediaUrl } from "../../utils/basePath";
import { parseDurationSeconds } from "../../utils/time";
import { TrainCard } from "../training/TrainingCard";
import { WorkoutPicker } from "../workouts/WorkoutPicker";
import { TrainingFinishModal } from "../training/FinishTrainModal";
import { TrainingOverrunModal } from "../training/OverrunTrainModal";

type OverrunState = {
  show: boolean;
  deadlineMs: number;
};

type OverrunRefState = {
  key: string | null;
  thresholdMs: number;
  postponedUntilMs: number | null;
  hasShown: boolean;
};

type SoundSchedule = {
  key: string | null;
  triggerAtMs: number;
  soundUrl: string;
};

const OVERRUN_GRACE_MS = 30_000;
const OVERRUN_MODAL_MS = 60_000;
const OVERRUN_COUNTDOWN_TICK_MS = 250;

// Train runs the active workout training.
export function TrainView({
  workouts,
  selectedWorkoutId,
  onSelectWorkout,
  onStartTrain,
  startDisabled,
  startTitle,
  training,
  currentStep,
  elapsed,
  workoutName,
  sounds,
  markSoundPlayed,
  onStartStep,
  onPause,
  onNext,
  onFinishTrain: onFinishTrain,
  onCopySummary,
  onToast,
  pauseOnTabHidden,
}: {
  workouts: Workout[];
  selectedWorkoutId: string | null;
  onSelectWorkout: (id: string) => void;
  onStartTrain: () => void | Promise<void>;
  startDisabled: boolean;
  startTitle?: string;
  training: TrainngState | null;
  currentStep: TrainngStepState | null;
  elapsed: number;
  workoutName: string;
  sounds: SoundOption[];
  markSoundPlayed: () => void;
  onStartStep: () => void;
  onPause: () => void;
  onNext: () => void;
  onFinishTrain: () => Promise<string | null>;
  onCopySummary: () => void;
  onToast: (message: string) => void;
  pauseOnTabHidden: boolean;
}) {
  const [finishSummary, setFinishSummary] = useState<string | null>(null);

  const [overrunModal, setOverrunModal] = useState<OverrunState | null>(null);
  const [overrunCountdown, setOverrunCountdown] = useState(0);

  const now = () => Date.now();

  // ---------- Refs for stable handlers ----------
  const trainingRef = useRef<TrainngState | null>(training);
  const currentStepRef = useRef<TrainngStepState | null>(currentStep);
  const elapsedRef = useRef(elapsed);
  const overrunModalRef = useRef<OverrunState | null>(overrunModal);

  useEffect(() => {
    trainingRef.current = training;
  }, [training]);
  useEffect(() => {
    currentStepRef.current = currentStep;
  }, [currentStep]);
  useEffect(() => {
    elapsedRef.current = elapsed;
  }, [elapsed]);
  useEffect(() => {
    overrunModalRef.current = overrunModal;
  }, [overrunModal]);

  // ---------- Audio pool ----------
  const activeAudiosRef = useRef<Set<HTMLAudioElement>>(new Set());
  const pausedAudiosRef = useRef<Set<HTMLAudioElement>>(new Set());
  const playTokenRef = useRef(0);

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

  const pauseActiveAudio = useCallback(() => {
    if (!activeAudiosRef.current.size) return;
    activeAudiosRef.current.forEach((audio) => {
      audio.pause();
      pausedAudiosRef.current.add(audio);
    });
    activeAudiosRef.current.clear();
  }, []);

  const resumePausedAudio = useCallback(() => {
    if (!pausedAudiosRef.current.size) return;
    pausedAudiosRef.current.forEach((audio) => {
      activeAudiosRef.current.add(audio);
      audio.play().catch(() => {
        activeAudiosRef.current.delete(audio);
      });
    });
    pausedAudiosRef.current.clear();
  }, []);

  const stopActiveAudio = useCallback(() => {
    playTokenRef.current += 1;

    if (!activeAudiosRef.current.size && !pausedAudiosRef.current.size) return;

    activeAudiosRef.current.forEach((audio) => stopAudio(audio));
    activeAudiosRef.current.clear();

    pausedAudiosRef.current.forEach((audio) => stopAudio(audio));
    pausedAudiosRef.current.clear();
  }, []);

  const handlePause = useCallback(() => {
    pauseActiveAudio();
    onPause();
  }, [onPause, pauseActiveAudio]);

  const handleStart = useCallback(() => {
    onStartStep();
    resumePausedAudio();
  }, [onStartStep, resumePausedAudio]);

  // ---------- Overrun scheduling ----------
  const overrunTimeoutRef = useRef<number | null>(null);
  const overrunIntervalRef = useRef<number | null>(null);

  const overrunRef = useRef<OverrunRefState>({
    key: null,
    thresholdMs: 0,
    postponedUntilMs: null,
    hasShown: false,
  });

  const clearOverrunTimers = useCallback(() => {
    if (overrunTimeoutRef.current) {
      clearTimeout(overrunTimeoutRef.current);
      overrunTimeoutRef.current = null;
    }
    if (overrunIntervalRef.current) {
      clearInterval(overrunIntervalRef.current);
      overrunIntervalRef.current = null;
    }
  }, []);

  const resetOverrunState = useCallback(() => {
    clearOverrunTimers();
    setOverrunModal(null);
    setOverrunCountdown(0);
  }, [clearOverrunTimers]);

  const handleOverrunPause = useCallback(() => {
    handlePause();
    resetOverrunState();
  }, [handlePause, resetOverrunState]);

  const handleOverrunPostpone = useCallback(() => {
    const s = trainingRef.current;
    if (!s?.running) return;

    overrunRef.current.postponedUntilMs = elapsedRef.current + OVERRUN_GRACE_MS;
    overrunRef.current.hasShown = false;
    resetOverrunState();
  }, [resetOverrunState]);

  useEffect(() => () => resetOverrunState(), [resetOverrunState]);

  useEffect(() => {
    if (!training?.running) resetOverrunState();
  }, [training?.running, resetOverrunState]);

  useEffect(() => {
    resetOverrunState();

    const estimateMs = currentStep?.estimatedSeconds
      ? currentStep.estimatedSeconds * 1000
      : 0;

    const key =
      training?.trainingId && typeof training.currentIndex === "number"
        ? `${training.trainingId}:${training.currentIndex}:${currentStep?.id || ""}`
        : null;

    overrunRef.current.key = key;
    overrunRef.current.thresholdMs =
      estimateMs > 0 ? estimateMs + OVERRUN_GRACE_MS : 0;
    overrunRef.current.postponedUntilMs = null;
    overrunRef.current.hasShown = false;
  }, [
    training?.trainingId,
    training?.currentIndex,
    currentStep?.id,
    currentStep?.estimatedSeconds,
    resetOverrunState,
  ]);

  useEffect(() => {
    if (!training?.running || !currentStep?.estimatedSeconds) return;
    if (overrunModal?.show) return;

    const thresholdMs = overrunRef.current.thresholdMs;
    if (thresholdMs <= 0) return;

    const postponedUntilMs = overrunRef.current.postponedUntilMs;
    const effectiveThreshold =
      postponedUntilMs && postponedUntilMs > thresholdMs
        ? postponedUntilMs
        : thresholdMs;

    if (elapsed < effectiveThreshold) return;
    if (overrunRef.current.hasShown) return;

    overrunRef.current.hasShown = true;

    const deadlineMs = now() + OVERRUN_MODAL_MS;
    setOverrunModal({ show: true, deadlineMs });
    setOverrunCountdown(OVERRUN_MODAL_MS);

    clearOverrunTimers();
    overrunTimeoutRef.current = window.setTimeout(() => {
      handleOverrunPause();
    }, OVERRUN_MODAL_MS);

    overrunIntervalRef.current = window.setInterval(() => {
      setOverrunCountdown(Math.max(0, deadlineMs - now()));
    }, OVERRUN_COUNTDOWN_TICK_MS);
  }, [
    training?.running,
    currentStep?.estimatedSeconds,
    elapsed,
    overrunModal?.show,
    clearOverrunTimers,
    handleOverrunPause,
  ]);

  useEffect(() => {
    if (!overrunModal?.show) {
      clearOverrunTimers();
      setOverrunCountdown(0);
      return;
    }
    setOverrunCountdown(Math.max(0, overrunModal.deadlineMs - now()));
    return () => clearOverrunTimers();
  }, [overrunModal?.show, overrunModal?.deadlineMs, clearOverrunTimers]);

  // ---------- Pause on tab hidden ----------
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
          pauseActiveAudio();
          onPause();
          if (!hiddenPauseNotifiedRef.current) {
            onToast("Training paused while tab was hidden");
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
  }, [pauseOnTabHidden, onPause, onToast, pauseActiveAudio]);

  // ---------- Per-training sound tracking ----------
  const trainingIdRef = useRef<string | null>(null);
  const subsetSoundPlayedRef = useRef(new Set<string>());

  useEffect(() => {
    const currentTrainingId = training?.trainingId || null;
    if (trainingIdRef.current !== currentTrainingId) {
      trainingIdRef.current = currentTrainingId;
      subsetSoundPlayedRef.current = new Set();
    }
  }, [training?.trainingId]);

  // ---------- Sound scheduling ----------
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

  useEffect(() => {
    return () => {
      clearSoundTimers();
      stopActiveAudio();
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
  }, [clearSoundTimers, stopActiveAudio]);

  // Schedule subset and exercise target sounds (NO timer logging here).
  useEffect(() => {
    if (!currentStep || !training?.running) {
      clearSoundTimers();
      if (!currentStep) stopActiveAudio();
      else pauseActiveAudio();
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

    const tokenAtSchedule = playTokenRef.current;

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
            if (playTokenRef.current !== tokenAtSchedule) return;

            stopActiveAudio();
            const audio = new Audio(subsetSoundUrl);
            activeAudiosRef.current.add(audio);
            audio.addEventListener(
              "ended",
              () => {
                activeAudiosRef.current.delete(audio);
                pausedAudiosRef.current.delete(audio);
              },
              { once: true },
            );
            audio.play().catch(() => {
              activeAudiosRef.current.delete(audio);
            });
            subsetSoundTimerRef.current = null;
            subsetSoundPlayedRef.current.add(subsetInstanceId);
          };

          if (remaining <= 0) fire();
          else {
            subsetSoundTimerRef.current = window.setTimeout(fire, remaining);
            subsetSoundScheduleRef.current.triggerAtMs = now() + remaining;
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
          if (playTokenRef.current !== tokenAtSchedule) return;

          stopActiveAudio();
          const audio = new Audio(exerciseSoundUrl);
          activeAudiosRef.current.add(audio);
          audio.addEventListener(
            "ended",
            () => {
              activeAudiosRef.current.delete(audio);
              pausedAudiosRef.current.delete(audio);
            },
            { once: true },
          );
          audio.play().catch(() => {
            activeAudiosRef.current.delete(audio);
          });
          stepSoundTimerRef.current = null;
          markSoundPlayed();
        };

        if (remaining <= 0) fire();
        else {
          stepSoundTimerRef.current = window.setTimeout(fire, remaining);
          stepSoundScheduleRef.current.triggerAtMs = now() + remaining;
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
    pauseActiveAudio,
    stopActiveAudio,
    clearSoundTimers,
  ]);

  // ---------- Keyboard shortcuts (ignore key repeat) ----------
  const runButtonRef = useRef<HTMLButtonElement>(null!);
  const nextActionButtonRef = useRef<HTMLButtonElement>(null!);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.repeat) return;

      const overrun = overrunModalRef.current;
      const s = trainingRef.current;

      if (overrun?.show) {
        if (e.code === "Enter") {
          e.preventDefault();
          handleOverrunPostpone();
        } else if (e.code === "Space") {
          e.preventDefault();
          handleOverrunPause();
        }
        return;
      }

      const tag = (e.target as HTMLElement | null)?.tagName;
      if (tag === "INPUT" || tag === "SELECT" || tag === "TEXTAREA") return;

      if (e.code === "Space") {
        e.preventDefault();
        if (!s) return;

        if (s.running) handlePause();
        else handleStart();
        return;
      }

      if (e.code === "Enter") {
        e.preventDefault();
        if (!s || s.done || !s.startedAt) return;

        const isLast =
          s.currentIndex >= (s.steps?.length ? s.steps.length - 1 : 0);

        stopActiveAudio();

        if (!isLast) {
          onNext();
          return;
        }

        onFinishTrain().then((summary) => {
          if (summary) setFinishSummary(summary);
        });
      }
    };

    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [
    handleOverrunPostpone,
    handleOverrunPause,
    handlePause,
    handleStart,
    stopActiveAudio,
    onNext,
    onFinishTrain,
  ]);

  const handleFinish = useCallback(async () => {
    const summary = await onFinishTrain();
    if (summary) setFinishSummary(summary);
  }, [onFinishTrain]);

  const headerStatus = useMemo(() => {
    if (!training) return null;

    const total = training.steps?.length || 0;
    const current =
      typeof training.currentIndex === "number" ? training.currentIndex + 1 : 0;

    if (training.done) return `Finished • ${total} steps`;
    if (!training.startedAt) return `Ready • ${total} steps`;
    if (training.running) return `Running • step ${current}/${total}`;
    return `Paused • step ${current}/${total}`;
  }, [training]);

  const startLabel =
    selectedWorkoutId && training?.workoutId === selectedWorkoutId
      ? "New"
      : "Select";

  return (
    <>
      <section className="panel">
        <div className="panel-header">
          <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
            <h3 style={{ margin: 0 }}>Train</h3>
            <div className="muted small">
              {headerStatus ? (
                <>
                  <strong>{workoutName || "Workout"}</strong>
                  {" • "}
                  {headerStatus}
                  {" • "}
                  {formatMillis(elapsed)}
                </>
              ) : (
                <span>Select a workout to start.</span>
              )}
            </div>
          </div>

          <div className="btn-group">
            <WorkoutPicker
              workouts={workouts}
              value={selectedWorkoutId}
              onSelect={onSelectWorkout}
              onClear={() => onSelectWorkout("")}
            />
            <button
              className="btn primary"
              onClick={onStartTrain}
              disabled={startDisabled}
              title={startTitle}
            >
              {startLabel}
            </button>
          </div>
        </div>

        <TrainCard
          training={training}
          currentStep={currentStep}
          elapsed={elapsed}
          workoutName={workoutName}
          onStart={handleStart}
          onPause={handlePause}
          onNext={() => {
            stopActiveAudio();
            onNext();
          }}
          onFinish={handleFinish}
          onStopAudio={stopActiveAudio}
          runButtonRef={runButtonRef}
          nextButtonRef={nextActionButtonRef}
        />
      </section>

      <TrainingFinishModal
        summary={finishSummary}
        onClose={() => setFinishSummary(null)}
        onCopySummary={onCopySummary}
      />
      <TrainingOverrunModal
        show={Boolean(overrunModal?.show)}
        countdown={overrunCountdown}
        onPause={handleOverrunPause}
        onPostpone={handleOverrunPostpone}
      />
    </>
  );
}
