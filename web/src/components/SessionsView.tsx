import { useCallback, useEffect, useRef, useState } from "react";

import type {
  SessionState,
  SessionStepState,
  SoundOption,
  Workout,
} from "../types";
import { formatMillis } from "../utils/format";
import { resolveMediaUrl } from "../utils/basePath";
import { parseDurationSeconds } from "../utils/time";
import { SessionCard } from "./SessionCard";

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
  const soundTimerRef = useRef<number | null>(null);
  const hiddenPauseNotifiedRef = useRef(false);

  // clearOverrunTimers stops any pending overrun timers.
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

  // handleOverrunPostpone delays the overrun modal by 30 seconds.
  const handleOverrunPostpone = useCallback(() => {
    if (!session?.running) return;
    nextOverrunMsRef.current = elapsed + 30000;
    setOverrunModal(null);
    setOverrunCountdown(0);
    clearOverrunTimers();
  }, [session?.running, elapsed, clearOverrunTimers]);

  // handleOverrunPause pauses the session from the overrun prompt.
  const handleOverrunPause = useCallback(() => {
    clearOverrunTimers();
    onPause();
    setOverrunModal(null);
    setOverrunCountdown(0);
  }, [onPause, clearOverrunTimers]);

  useEffect(() => () => clearOverrunTimers(), [clearOverrunTimers]);

  // Cleanup any pending sound timer on unmount.
  useEffect(() => {
    return () => {
      if (soundTimerRef.current) {
        clearTimeout(soundTimerRef.current);
        soundTimerRef.current = null;
      }
    };
  }, []);

  useEffect(() => {
    if (!session?.running) {
      clearOverrunTimers();
      setOverrunModal(null);
      setOverrunCountdown(0);
    }
  }, [session?.running, clearOverrunTimers]);

  useEffect(() => {
    const handleVisibility = () => {
      if (document.hidden) {
        if (session?.running) {
          onPause();
          if (!hiddenPauseNotifiedRef.current) {
            onToast("Session paused while tab was hidden");
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
    // Reset overrun scheduling when step/session changes.
    clearOverrunTimers();
    setOverrunModal(null);
    setOverrunCountdown(0);
    const estimateMs = currentStep?.estimatedSeconds
      ? currentStep.estimatedSeconds * 1000
      : null;
    nextOverrunMsRef.current = estimateMs ? estimateMs + 30000 : null;
  }, [
    session?.sessionId,
    session?.currentIndex,
    currentStep?.id,
    clearOverrunTimers,
    currentStep?.estimatedSeconds,
  ]);

  useEffect(() => {
    if (soundTimerRef.current) {
      clearTimeout(soundTimerRef.current);
      soundTimerRef.current = null;
    }
    // Guard: nothing to play when no active, running step.
    if (!currentStep || !session?.running) return;
    // Guard: only fire once per step.
    if (currentStep.soundPlayed) return;
    // Guard: skip when no sound is configured.
    if (!currentStep.soundKey && !currentStep.soundUrl) return;

    const soundOpt =
      sounds.find((s) => s.key === currentStep.soundKey) ||
      sounds.find((s) => s.file === currentStep.soundUrl);
    // Wait until we know the sound to preserve leadSeconds.
    if (currentStep.soundKey && !soundOpt && !currentStep.soundUrl) return;

    const soundUrl = resolveMediaUrl(
      currentStep.soundUrl || soundOpt?.file || "",
    );
    if (!soundUrl) return;

    const baseLeadSeconds = soundOpt?.leadSeconds ?? 0;
    const targetSeconds =
      currentStep.estimatedSeconds ||
      parseDurationSeconds((currentStep as any).duration);
    const targetMs = targetSeconds * 1000;
    if (targetMs <= 0) return;

    const leadMs = Math.min(
      targetMs,
      Math.max(0, (baseLeadSeconds || 0) * 1000),
    );
    const triggerMs = Math.max(0, targetMs - leadMs);
    const remaining = triggerMs - elapsed;

    // play triggers the sound once and marks the step as played.
    const play = () => {
      new Audio(soundUrl).play().catch(() => {});
      markSoundPlayed();
    };

    if (remaining <= 0) {
      play();
      return;
    }

    soundTimerRef.current = window.setTimeout(play, remaining);
    return () => {
      if (soundTimerRef.current) {
        clearTimeout(soundTimerRef.current);
        soundTimerRef.current = null;
      }
    };
  }, [
    currentStep?.id,
    currentStep?.soundKey,
    currentStep?.soundUrl,
    currentStep?.estimatedSeconds,
    (currentStep as any)?.duration,
    session?.sessionId,
    session?.currentIndex,
    session?.running,
    sounds,
    markSoundPlayed,
    elapsed,
  ]);

  useEffect(() => {
    // Show overrun modal when elapsed passes target + 30s.
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
    // Maintain countdown while overrun modal is visible.
    if (!overrunModal?.show) {
      clearOverrunTimers();
      setOverrunCountdown(0);
      return;
    }
    setOverrunCountdown(Math.max(0, (overrunModal.deadline || 0) - Date.now()));
    return () => clearOverrunTimers();
  }, [overrunModal?.show, overrunModal?.deadline, clearOverrunTimers]);

  useEffect(() => {
    // Keyboard shortcuts for pause/resume and next/finish.
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
        if (session.running) {
          onPause();
          return;
        }
        onStartStep();
      }
      if (e.code === "Enter") {
        e.preventDefault();
        if (!session || session.done || !session.startedAt) return;
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
  ]);

  // handleFinish triggers session completion and opens the summary modal.
  const handleFinish = async () => {
    const summary = await onFinishSession();
    if (summary) setFinishSummary(summary);
  };

  return (
    <>
      <section className="panel">
        <div className="panel-header">
          <div>
            <p className="label">Workout</p>
            <select
              value={selectedWorkoutId || ""}
              onChange={(e) => onSelectWorkout(e.target.value)}
            >
              <option value="">Select</option>
              {workouts.map((w) => (
                <option key={w.id} value={w.id}>
                  {w.name} ({w.steps.length})
                </option>
              ))}
            </select>
          </div>
          <button
            className="btn primary"
            onClick={onStartSession}
            disabled={startDisabled}
            title={startTitle}
          >
            Start Session
          </button>
        </div>
        <SessionCard
          session={session}
          currentStep={currentStep}
          elapsed={elapsed}
          workoutName={workoutName}
          onStart={onStartStep}
          onPause={onPause}
          onNext={onNext}
          onFinish={handleFinish}
        />
      </section>
      {finishSummary && (
        <div className="modal-overlay" onClick={() => setFinishSummary(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h3>Great job!</h3>
            <p className="muted">Session finished. Copy the summary for AI.</p>
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
