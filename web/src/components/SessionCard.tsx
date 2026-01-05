import { useEffect, useMemo, useState } from "react";
import {
  formatExerciseLine,
  formatExercises,
  formatMillis,
} from "../utils/format";
import type { Exercise, SessionState, WorkoutStep } from "../types";

// SessionCard shows the live session with controls and steps.
export function SessionCard({
  session,
  currentStep,
  elapsed,
  workoutName,
  onStart,
  onPause,
  onNext,
  onFinish,
}: {
  session: SessionState | null;
  currentStep: WorkoutStep | null;
  elapsed: number;
  workoutName?: string;
  onStart: () => void;
  onPause: () => void;
  onNext: () => void;
  onFinish: () => void;
}) {
  const running = session?.running;
  const done = session?.done;
  const [expandedSessionSteps, setExpandedSessionSteps] = useState<Set<string>>(
    new Set(),
  );
  const nextStep = useMemo(() => {
    if (!session) return null;
    return session.steps.find(
      (s, idx) => idx > session.currentIndex && !s.completed,
    );
  }, [session]);
  const currentExercises = useMemo(() => {
    if (!currentStep) return [];
    if ((currentStep as any).exercises?.length)
      return (currentStep as any).exercises;
    const fallback = {
      name: (currentStep as any).exercise || "",
      amount: (currentStep as any).amount,
      weight: (currentStep as any).weight,
    };
    return fallback.name || fallback.amount || fallback.weight
      ? [fallback]
      : [];
  }, [currentStep]);
  const isLastStep =
    session && session.steps.length
      ? session.currentIndex >= session.steps.length - 1
      : false;
  const hasProgress = session?.steps?.some(
    (s) => s.elapsedMillis > 0 || s.completed,
  );
  const hasStarted = Boolean(session?.running) || Boolean(hasProgress);
  // Decide primary button label based on running state.
  const startLabel = session?.running
    ? "Pause"
    : hasStarted
      ? "Continue"
      : "Start";
  const displayMillis =
    currentStep &&
    currentStep.type === "pause" &&
    (currentStep as any).pauseOptions?.autoAdvance &&
    currentStep.estimatedSeconds
      ? Math.max(
          0,
          currentStep.estimatedSeconds * 1000 - Math.max(0, elapsed - 1000), // add a extra second so the start time is accurate, otherwise you loose the starting second
        )
      : elapsed;
  const totalSteps = session?.steps?.length || 0;
  const currentNumber = session ? session.currentIndex + 1 : 0;

  useEffect(() => {
    // Auto-advance when a pause with autoAdvance reaches zero.
    if (
      !session ||
      !running ||
      !currentStep ||
      currentStep.type !== "pause" ||
      !(currentStep as any).pauseOptions?.autoAdvance ||
      !currentStep.estimatedSeconds
    ) {
      return;
    }
    const remaining = currentStep.estimatedSeconds * 1000 - elapsed;
    if (remaining <= 0) {
      onNext();
      return;
    }
    const timer = setTimeout(() => {
      onNext();
    }, remaining);
    return () => clearTimeout(timer);
  }, [
    session?.sessionId,
    session?.currentIndex,
    running,
    currentStep,
    elapsed,
    onNext,
  ]);

  // Reset expansions when session changes.
  useEffect(() => {
    if (!session) {
      setExpandedSessionSteps(new Set());
      return;
    }
    setExpandedSessionSteps(new Set());
  }, [session?.sessionId]);

  // Collapse when moving to a new current index.
  useEffect(() => {
    if (!session) return;
    setExpandedSessionSteps(new Set());
  }, [session?.currentIndex]);

  return (
    <div className="session-card">
      <div className="session-main">
        <div className="current-card">
          <div className="label muted">Now</div>
          {workoutName ? (
            <div className="muted small">{workoutName}</div>
          ) : null}
          {session ? (
            <div className="muted small">
              Step {currentNumber}/{totalSteps}
            </div>
          ) : null}
          <div className="clock-row">
            <div className="clock">
              {currentStep ? formatMillis(displayMillis) : "00:00"}
            </div>
          </div>
          <div className="current-step">
            {currentStep ? currentStep.name : "No session"}
          </div>
          {currentExercises.length > 0 && (
            <div className="exercise-pills">
              {currentExercises.map((ex: Exercise, idx: number) => {
                const text = formatExerciseLine(ex);
                if (!text) return null;
                return (
                  <span key={idx} className="pill">
                    {text}
                  </span>
                );
              })}
            </div>
          )}
        </div>
        <div className="next-card">
          <div className="label muted">Next</div>
          <div className="next-name">{nextStep ? nextStep.name : "None"}</div>
          <div className="muted small">
            {nextStep
              ? (nextStep as any).estimatedSeconds
                ? `${(nextStep as any).estimatedSeconds}s`
                : (nextStep as any).duration || ""
              : ""}
          </div>
          {/* Primary session controls */}
          <div className="actions vertical session-actions">
            <button
              className="btn primary"
              onClick={running ? onPause : onStart}
              disabled={!session || done}
            >
              {startLabel}
            </button>
            <button
              className="btn large next"
              onClick={isLastStep ? onFinish : onNext}
              disabled={!session || done || !session.startedAt}
            >
              {isLastStep ? "Finish" : "Next"}
            </button>
          </div>
          <div
            className="muted small shortcuts"
            title="Space = pause/resume, Enter = next/finish"
          >
            Shortcuts ⓘ
          </div>
        </div>
      </div>
      <div className="session-steps">
        {/* Active + upcoming steps */}
        {session?.steps
          .filter((step: any) => step.current || !step.completed)
          .map((step: any, idx: number) => {
            const stepId = step.id || `step-${idx}`;
            const expanded = expandedSessionSteps.has(stepId);
            const originalIndex =
              session?.steps?.findIndex((s) => s.id === step.id) ?? idx;
            const displayIndex =
              originalIndex >= 0 ? originalIndex + 1 : idx + 1;
            return (
              <div
                key={stepId}
                className={
                  step.current
                    ? "step-row current"
                    : step.completed
                      ? "step-row done"
                      : "step-row"
                }
                onClick={() => {
                  if (step.type === "pause" && !formatExercises(step)) return;
                  setExpandedSessionSteps((prev) => {
                    const next = new Set(prev);
                    if (next.has(stepId)) next.delete(stepId);
                    else next.add(stepId);
                    return next;
                  });
                }}
              >
                <div className="badge small">{displayIndex}</div>
                <span className="chevron">
                  {step.type === "pause" && !formatExercises(step)
                    ? ""
                    : expanded
                      ? "▾"
                      : "▸"}
                </span>
                <div className="step-info">
                  <strong>{step.name}</strong>
                  <div className="muted">
                    {step.type} •{" "}
                    {step.estimatedSeconds
                      ? `${step.estimatedSeconds}s target`
                      : step.duration || "open"}{" "}
                  </div>
                  {expanded && formatExercises(step) && (
                    <div className="muted small">{formatExercises(step)}</div>
                  )}
                </div>
              </div>
            );
          })}
        {!session && <p className="muted">Start a session to see steps.</p>}
      </div>
    </div>
  );
}
