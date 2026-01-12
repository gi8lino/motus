import { useCallback, useEffect, useRef, useState, useMemo } from "react";

import type {
  SessionState,
  SessionStepState,
  SoundOption,
  Workout,
} from "../types";
import { formatMillis } from "../utils/format";
import { resolveMediaUrl } from "../utils/basePath";
import { parseDurationSeconds } from "../utils/time";
import { logTimerEvent } from "../utils/timerLogger";
import { SessionCard } from "./SessionCard";
import { WorkoutSelect } from "./WorkoutSelect";

// SessionsView runs the active workout session.
export function SessionsView({
  workouts,
  selectedWorkoutId,
  onSelectWorkout,
  onStartSession,
  startDisabled,
  startTitle,
  session,
  currentStep,
  elapsed,
  workoutName,
  sounds,
  markSoundPlayed,
  onStartStep,
  onPause,
  onNext,
  onFinishSession,
  onCopySummary,
  onToast,
}: {
  workouts: Workout[];
  selectedWorkoutId: string | null;
  onSelectWorkout: (id: string) => void;
  onStartSession: () => void | Promise<void>;
  startDisabled: boolean;
  startTitle?: string;
  session: SessionState | null;
  currentStep: SessionStepState | null;
  elapsed: number;
  workoutName: string;
  sounds: SoundOption[];
  markSoundPlayed: () => void;
  onStartStep: () => void;
  onPause: () => void;
  onNext: () => void;
  onFinishSession: () => Promise<string | null>;
  onCopySummary: () => void;
  onToast: (message: string) => void;
}) {
  const [finishSummary, setFinishSummary] = useState<string | null>(null);
  const [overrunModal, setOverrunModal] = useState<{
    show: boolean;
    deadline: number;
  } | null>(null);
  const [overrunCountdown, setOverrunCountdown] = useState(0);
  const nextOverrunMsRef = useRef<number | null>(null);
  const overrunTimeoutRef = useRef<number | null>(null);
  const overrunIntervalRef = useRef<number | null>(null);
  const stepSoundTimerRef = useRef<number | null>(null);
  const subsetSoundTimerRef = useRef<number | null>(null);
  const stepSoundScheduleRef = useRef<{
    key: string | null;
    targetMs: number;
    soundUrl: string;
    triggerAt: number;
  }>({ key: null, targetMs: 0, soundUrl: "", triggerAt: 0 });
  const subsetSoundScheduleRef = useRef<{
    key: string | null;
    targetMs: number;
    soundUrl: string;
    triggerAt: number;
  }>({ key: null, targetMs: 0, soundUrl: "", triggerAt: 0 });
  const subsetSoundPlayedRef = useRef(new Set<string>());
  const sessionIdRef = useRef<string | null>(null);
  const hiddenPauseNotifiedRef = useRef(false);
  const keyCooldownRef = useRef<Record<string, number>>({});
  const buttonCooldownTimersRef = useRef<Record<string, number | null>>({});
  const elapsedRef = useRef(0);
  const runButtonRef = useRef<HTMLButtonElement>(null!);
  const nextActionButtonRef = useRef<HTMLButtonElement>(null!);

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

  const handleOverrunPostpone = useCallback(() => {
    if (!session?.running) return;
    nextOverrunMsRef.current = elapsed + 30000;
    resetOverrunState();
  }, [session?.running, elapsed, resetOverrunState]);

  const handleOverrunPause = useCallback(() => {
    onPause();
    resetOverrunState();
  }, [onPause, resetOverrunState]);

  useEffect(() => () => resetOverrunState(), [resetOverrunState]);

  useEffect(() => {
    const currentSessionId = session?.sessionId || null;
    if (sessionIdRef.current !== currentSessionId) {
      sessionIdRef.current = currentSessionId;
      subsetSoundPlayedRef.current = new Set();
    }
  }, [session?.sessionId]);

  useEffect(() => {
    elapsedRef.current = elapsed;
  }, [elapsed]);

  const tryConsumeKey = useCallback(
    (code: string, button?: HTMLButtonElement | null) => {
      const now = Date.now();
      const last = keyCooldownRef.current[code];
      if (last && now - last < 2000) {
        return false;
      }
      keyCooldownRef.current[code] = now;
      if (button) {
        button.classList.add("key-cooldown");
        const existing = buttonCooldownTimersRef.current[code];
        if (existing) {
          window.clearTimeout(existing);
        }
        buttonCooldownTimersRef.current[code] = window.setTimeout(() => {
          button.classList.remove("key-cooldown");
          buttonCooldownTimersRef.current[code] = null;
        }, 2000);
      }
      return true;
    },
    [],
  );

  useEffect(() => {
    return () => {
      Object.values(buttonCooldownTimersRef.current).forEach((timer) => {
        if (timer) {
          window.clearTimeout(timer);
        }
      });
    };
  }, []);

  useEffect(() => {
    return () => {
      if (stepSoundTimerRef.current) {
        clearTimeout(stepSoundTimerRef.current);
        stepSoundTimerRef.current = null;
      }
      if (subsetSoundTimerRef.current) {
        clearTimeout(subsetSoundTimerRef.current);
        subsetSoundTimerRef.current = null;
      }
      stepSoundScheduleRef.current = {
        key: null,
        targetMs: 0,
        soundUrl: "",
        triggerAt: 0,
      };
      subsetSoundScheduleRef.current = {
        key: null,
        targetMs: 0,
        soundUrl: "",
        triggerAt: 0,
      };
    };
  }, []);

  useEffect(() => {
    if (!session?.running) {
      resetOverrunState();
    }
  }, [session?.running, resetOverrunState]);

  useEffect(() => {
    const handleVisibility = () => {
      if (document.hidden) {
        if (session?.running) {
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
  }, [session?.running, onPause, onToast]);

  useEffect(() => {
    resetOverrunState();
    const estimateMs = currentStep?.estimatedSeconds
      ? currentStep.estimatedSeconds * 1000
      : null;
    nextOverrunMsRef.current = estimateMs ? estimateMs + 30000 : null;
  }, [
    session?.sessionId,
    session?.currentIndex,
    currentStep?.id,
    resetOverrunState,
    currentStep?.estimatedSeconds,
  ]);

  useEffect(() => {
    if (!currentStep || !session?.running) {
      if (stepSoundTimerRef.current) {
        clearTimeout(stepSoundTimerRef.current);
        stepSoundTimerRef.current = null;
      }
      if (subsetSoundTimerRef.current) {
        clearTimeout(subsetSoundTimerRef.current);
        subsetSoundTimerRef.current = null;
      }
      stepSoundScheduleRef.current = {
        key: null,
        targetMs: 0,
        soundUrl: "",
        triggerAt: 0,
      };
      subsetSoundScheduleRef.current = {
        key: null,
        targetMs: 0,
        soundUrl: "",
        triggerAt: 0,
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

    if (hasSubsetTarget && currentStep.subsetId && subsetSoundUrl) {
      const subsetInstanceId = `${currentStep.subsetId}-${
        currentStep.loopIndex ?? 0
      }`;
      if (!subsetSoundPlayedRef.current.has(subsetInstanceId)) {
        const subsetTargetMs = subsetTargetSeconds * 1000;
        const leadMs = Math.min(
          subsetTargetMs,
          Math.max(0, subsetLeadSeconds * 1000),
        );
        const triggerMs = Math.max(0, subsetTargetMs - leadMs);
        const schedule = subsetSoundScheduleRef.current;
        if (
          !subsetSoundTimerRef.current ||
          schedule.key !== subsetInstanceId ||
          schedule.targetMs !== subsetTargetMs ||
          schedule.soundUrl !== subsetSoundUrl
        ) {
          const subsetLogBase = {
            sessionId: session?.sessionId,
            stepId: currentStep.id,
            subsetId: currentStep.subsetId,
            loopIndex: currentStep.loopIndex ?? 0,
            targetMs: subsetTargetMs,
            soundUrl: subsetSoundUrl,
          };
          if (subsetSoundTimerRef.current) {
            logTimerEvent("subset-sound-cancelled", subsetLogBase);
            clearTimeout(subsetSoundTimerRef.current);
            subsetSoundTimerRef.current = null;
          }
          subsetSoundScheduleRef.current = {
            key: subsetInstanceId,
            targetMs: subsetTargetMs,
            soundUrl: subsetSoundUrl,
            triggerAt: 0,
          };
          const subsetElapsedMs = (() => {
            const currentElapsed = elapsedRef.current;
            const targetLoop = currentStep.loopIndex ?? 0;
            return session.steps.reduce((acc, step, idx) => {
              if (step.subsetId !== currentStep.subsetId) return acc;
              const stepLoop = step.loopIndex ?? 0;
              if (stepLoop !== targetLoop) return acc;
              if (idx < session.currentIndex) {
                return acc + (step.elapsedMillis || 0);
              }
              if (idx === session.currentIndex) {
                return acc + currentElapsed;
              }
              return acc;
            }, 0);
          })();
          const remaining = triggerMs - subsetElapsedMs;
          const triggerAt = Date.now() + Math.max(remaining, 0);
          const play = () => {
            logTimerEvent("subset-sound-fired", subsetLogBase);
            new Audio(subsetSoundUrl).play().catch(() => {});
            subsetSoundTimerRef.current = null;
            subsetSoundPlayedRef.current.add(subsetInstanceId);
          };
          if (remaining <= 0) {
            play();
          } else {
            subsetSoundTimerRef.current = window.setTimeout(play, remaining);
            subsetSoundScheduleRef.current.triggerAt = triggerAt;
            logTimerEvent("subset-sound-scheduled", {
              ...subsetLogBase,
              remainingMs: remaining,
              triggerAt,
            });
          }
        }
      }
    }

    if (allowStepSound && exerciseTargetSeconds > 0 && exerciseSoundUrl) {
      const stepTargetMs = exerciseTargetSeconds * 1000;
      const leadMs = Math.min(
        stepTargetMs,
        Math.max(0, exerciseLeadSeconds * 1000),
      );
      const triggerMs = Math.max(0, stepTargetMs - leadMs);
      const scheduleKey =
        currentStep.id || `${session.sessionId}-${session.currentIndex}`;
      const schedule = stepSoundScheduleRef.current;
      const nowTs = Date.now();
      const stepLogBase = {
        sessionId: session?.sessionId,
        stepId: currentStep.id,
        soundUrl: exerciseSoundUrl,
        targetMs: stepTargetMs,
        loopIndex: currentStep.loopIndex ?? 0,
      };
      if (
        !stepSoundTimerRef.current ||
        schedule.key !== scheduleKey ||
        schedule.targetMs !== stepTargetMs ||
        schedule.soundUrl !== exerciseSoundUrl
      ) {
        if (stepSoundTimerRef.current) {
          const shouldKeep =
            schedule.key &&
            schedule.key !== scheduleKey &&
            schedule.triggerAt > 0 &&
            schedule.triggerAt <= nowTs + 150;
          if (!shouldKeep) {
            logTimerEvent("step-sound-cancelled", stepLogBase);
            clearTimeout(stepSoundTimerRef.current);
            stepSoundTimerRef.current = null;
          }
        }
        stepSoundScheduleRef.current = {
          key: scheduleKey,
          targetMs: stepTargetMs,
          soundUrl: exerciseSoundUrl,
          triggerAt: 0,
        };
        const remaining = triggerMs - elapsedRef.current;
        const triggerAt = nowTs + Math.max(remaining, 0);
        const play = () => {
          logTimerEvent("step-sound-fired", stepLogBase);
          new Audio(exerciseSoundUrl).play().catch(() => {});
          stepSoundTimerRef.current = null;
          markSoundPlayed();
        };
        if (remaining <= 0) {
          play();
        } else {
          stepSoundTimerRef.current = window.setTimeout(() => {
            play();
          }, remaining);
          stepSoundScheduleRef.current.triggerAt = triggerAt;
          logTimerEvent("step-sound-scheduled", {
            ...stepLogBase,
            remainingMs: remaining,
            triggerAt,
          });
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
    session?.sessionId,
    session?.currentIndex,
    session?.running,
    sounds,
    markSoundPlayed,
  ]);

  useEffect(() => {
    if (
      !session?.running ||
      !currentStep?.estimatedSeconds ||
      !nextOverrunMsRef.current
    ) {
      return;
    }
    if (overrunModal?.show) return;
    if (elapsed < nextOverrunMsRef.current) return;

    const deadline = Date.now() + 60000;
    setOverrunModal({ show: true, deadline });
    setOverrunCountdown(60000);
    clearOverrunTimers();
    overrunTimeoutRef.current = window.setTimeout(() => {
      handleOverrunPause();
    }, 60000);
    overrunIntervalRef.current = window.setInterval(() => {
      setOverrunCountdown(Math.max(0, deadline - Date.now()));
    }, 250);
  }, [
    session?.running,
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
    setOverrunCountdown(Math.max(0, (overrunModal.deadline || 0) - Date.now()));
    return () => clearOverrunTimers();
  }, [overrunModal?.show, overrunModal?.deadline, clearOverrunTimers]);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (overrunModal?.show) {
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
        if (!session) return;
        if (!tryConsumeKey(e.code, runButtonRef.current)) return;
        if (session.running) {
          onPause();
          return;
        }
        onStartStep();
        return;
      }

      if (e.code === "Enter") {
        e.preventDefault();
        if (!session || session.done || !session.startedAt) return;
        if (!tryConsumeKey(e.code, nextActionButtonRef.current)) return;
        const isLast =
          session.currentIndex >=
          (session.steps?.length ? session.steps.length - 1 : 0);
        if (!isLast) {
          onNext();
          return;
        }
        onFinishSession().then((summary) => {
          if (summary) setFinishSummary(summary);
        });
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [
    session,
    overrunModal?.show,
    handleOverrunPause,
    handleOverrunPostpone,
    onStartStep,
    onPause,
    onNext,
    onFinishSession,
    tryConsumeKey,
  ]);

  const handleFinish = async () => {
    const summary = await onFinishSession();
    if (summary) setFinishSummary(summary);
  };

  const headerStatus = useMemo(() => {
    if (!session) return null;

    const total = session.steps?.length || 0;
    const current =
      typeof session.currentIndex === "number" ? session.currentIndex + 1 : 0;

    if (session.done) return `Finished • ${total} steps`;
    if (!session.startedAt) return `Ready • ${total} steps`;
    if (session.running) return `Running • step ${current}/${total}`;
    return `Paused • step ${current}/${total}`;
  }, [session]);

  const startLabel =
    selectedWorkoutId && session?.workoutId === selectedWorkoutId
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
            <WorkoutSelect
              workouts={workouts}
              value={selectedWorkoutId}
              onSelect={onSelectWorkout}
              onClear={() => onSelectWorkout("")}
            />
            <button
              className="btn primary"
              onClick={onStartSession}
              disabled={startDisabled}
              title={startTitle}
            >
              {startLabel}
            </button>
          </div>
        </div>

        <SessionCard
          session={session}
          currentStep={currentStep}
          elapsed={elapsed}
          onStart={onStartStep}
          onPause={onPause}
          onNext={onNext}
          onFinish={handleFinish}
          runButtonRef={runButtonRef}
          nextButtonRef={nextActionButtonRef}
        />
      </section>

      {finishSummary && (
        <div className="modal-overlay" onClick={() => setFinishSummary(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>Great job!</h3>
            <p className="muted">Training finished. Copy the summary for AI.</p>
            <textarea
              readOnly
              value={finishSummary}
              style={{ width: "100%", minHeight: "180px" }}
            />
            <div className="btn-group" style={{ justifyContent: "flex-end" }}>
              <button
                className="btn subtle"
                onClick={() => {
                  if (navigator?.clipboard?.writeText) {
                    navigator.clipboard
                      .writeText(finishSummary)
                      .catch(() => {});
                  }
                  onCopySummary();
                }}
              >
                Copy
              </button>
              <button
                className="btn primary"
                onClick={() => setFinishSummary(null)}
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}

      {overrunModal?.show && (
        <div className="modal-overlay" onClick={handleOverrunPause}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>Still training?</h3>
            <p className="muted">
              You passed the target. Auto-pause in{" "}
              {formatMillis(overrunCountdown)}.
            </p>
            <div className="btn-group" style={{ justifyContent: "flex-end" }}>
              <button className="btn subtle" onClick={handleOverrunPostpone}>
                Postpone (+30s)
              </button>
              <button className="btn primary" onClick={handleOverrunPause}>
                Pause
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
