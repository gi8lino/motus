import { useCallback, useEffect, useMemo, type RefObject } from "react";
import { formatExerciseLine, formatMillis } from "../../utils/format";
import { STEP_TYPE_PAUSE } from "../../utils/step";
import type { Exercise, SessionState, SessionStepState } from "../../types";
import { logTimerEvent } from "../../utils/timerLogger";

type AnyStep = any;

// getExercises normalizes the exercises list for a step payload.
function getExercises(step: AnyStep): Exercise[] {
  if (!step) return [];
  if (Array.isArray(step.exercises)) return step.exercises;
  return [];
}

// getStepName resolves the display label for a step.
function getStepName(step: AnyStep): string {
  const subsetLabel = String(step?.subsetLabel || "").trim();
  if (subsetLabel) return subsetLabel;
  const parent = String(step?.setName || "").trim();
  if (parent) return parent;
  const name = String(step?.name || "").trim();
  if (name) return name;
  if (step?.type === STEP_TYPE_PAUSE) return "Pause";
  return "Set";
}

// getCurrentExerciseLabel builds the display label for the active exercise.
function getCurrentExerciseLabel(step: AnyStep): string {
  if (!step) return "";
  const exercise = Array.isArray(step.exercises)
    ? step.exercises[0]
    : undefined;
  const formatted = exercise ? formatExerciseLine(exercise) : "";
  return formatted || getStepName(step);
}
type SubsetDisplay = {
  key: string;
  superset: boolean;
  label?: string;
  exercises: Exercise[];
};

type StepGroup = {
  key: string;
  setName: string;
  type: string;
  loopIndex: number;
  loopTotal: number;
  current: boolean;
  subsets: SubsetDisplay[];
  hasSuperset: boolean;
  estimatedSeconds: number;
};

// SessionCard renders the main session status card and controls.
export function SessionCard({
  session,
  currentStep,
  elapsed,
  workoutName,
  onStart,
  onPause,
  onNext,
  onFinish,
  runButtonRef,
  nextButtonRef,
}: {
  session: SessionState | null;
  currentStep: SessionStepState | null;
  elapsed: number;
  workoutName?: string;
  onStart: () => void;
  onPause: () => void;
  onNext: () => void;
  onFinish: () => void;
  runButtonRef?: RefObject<HTMLButtonElement>;
  nextButtonRef?: RefObject<HTMLButtonElement>;
}) {
  const running = session?.running;
  const done = session?.done;

  const isLastStep =
    session && session.steps.length
      ? session.currentIndex >= session.steps.length - 1
      : false;

  const hasProgress = session?.steps?.some(
    (s: any) => s.elapsedMillis > 0 || s.completed,
  );
  const hasStarted = Boolean(session?.running) || Boolean(hasProgress);

  const isAutoAdvance =
    (currentStep?.type === STEP_TYPE_PAUSE &&
      (currentStep as any).pauseOptions?.autoAdvance) ||
    Boolean((currentStep as any)?.autoAdvance);

  const startLabel = session?.running
    ? "Pause"
    : hasStarted
      ? "Continue"
      : "Start";

  const displayMillis =
    currentStep && isAutoAdvance && currentStep.estimatedSeconds
      ? Math.max(
          0,
          currentStep.estimatedSeconds * 1000 - Math.max(0, elapsed - 1000),
        )
      : elapsed;

  const currentExerciseLabel = useMemo(() => {
    if (!currentStep) return "No training";
    return getCurrentExerciseLabel(currentStep as any);
  }, [currentStep]);

  const extractExerciseLabels = useCallback(
    (step: SessionStepState | null, startIndex = 0) => {
      if (!step) return [];
      if (step.type === STEP_TYPE_PAUSE) {
        const durationText = step.estimatedSeconds
          ? formatMillis(step.estimatedSeconds * 1000)
          : "";
        const pauseText = durationText ? `Pause • ${durationText}` : "Pause";
        return [pauseText];
      }

      if (step.superset && step.subsetId && session?.steps?.length) {
        const subsetId = step.subsetId;
        const seen = new Set<string>();
        const labels: string[] = [];
        for (let idx = startIndex; idx < session.steps.length; idx += 1) {
          const candidate = session.steps[idx];
          if (candidate.subsetId !== subsetId) continue;
          for (const ex of getExercises(candidate)) {
            const text = formatExerciseLine(ex);
            if (text && !seen.has(text)) {
              seen.add(text);
              labels.push(text);
            }
          }
        }
        if (labels.length) return labels;
      }

      return getExercises(step)
        .map((ex) => formatExerciseLine(ex))
        .filter(Boolean);
    },
    [session?.steps],
  );

  const currentStepPills = useMemo(
    () => extractExerciseLabels(currentStep),
    [currentStep, extractExerciseLabels],
  );

  useEffect(() => {
    // Auto-advance when a timed exercise or pause reaches zero.
    if (
      !session ||
      !running ||
      !currentStep ||
      !isAutoAdvance ||
      !currentStep.estimatedSeconds
    ) {
      return;
    }
    const remaining = currentStep.estimatedSeconds * 1000 - elapsed;
    const autoAdvanceDetails = {
      sessionId: session?.sessionId,
      currentIndex: session?.currentIndex,
      stepId: currentStep?.id,
      remainingMs: remaining,
    };
    if (remaining <= 0) {
      logTimerEvent("auto-advance-step", {
        ...autoAdvanceDetails,
        triggered: true,
      });
      onNext();
      return;
    }
    logTimerEvent("auto-advance-step", {
      ...autoAdvanceDetails,
      scheduledInMs: remaining,
    });
    const timer = setTimeout(() => {
      logTimerEvent("auto-advance-step", {
        ...autoAdvanceDetails,
        triggered: true,
      });
      onNext();
    }, remaining);
    return () => clearTimeout(timer);
  }, [
    session?.sessionId,
    session?.currentIndex,
    running,
    currentStep,
    isAutoAdvance,
    elapsed,
    onNext,
  ]);

  const totalSteps = session?.steps?.length || 0;
  const currentNumber = session ? session.currentIndex + 1 : 0;

  const remainingSteps = useMemo(
    () => (session?.steps || []).filter((s: any) => !s.completed),
    [session?.sessionId, session?.steps],
  );

  const groupedSteps = useMemo(() => {
    if (!remainingSteps.length) return [];
    const groups: StepGroup[] = [];
    let currentGroup: StepGroup | null = null;
    for (const step of remainingSteps) {
      const setName =
        String(step?.setName || step?.name || "").trim() || "Step";
      const loopIndex = step?.loopIndex ?? 0;
      const loopTotal = step?.loopTotal ?? 0;
      const type = step?.type || "set";
      const needsNewGroup =
        !currentGroup ||
        currentGroup.setName !== setName ||
        currentGroup.loopIndex !== loopIndex ||
        currentGroup.loopTotal !== loopTotal ||
        currentGroup.type !== type;

      if (needsNewGroup) {
        currentGroup = {
          key: `${setName}-${loopIndex}-${loopTotal}-${groups.length}`,
          setName,
          type,
          loopIndex,
          loopTotal,
          current: Boolean(step.current),
          subsets: [],
          hasSuperset: false,
          estimatedSeconds: step?.estimatedSeconds ?? 0,
        };
        groups.push(currentGroup);
      } else if (currentGroup) {
        currentGroup.current = currentGroup.current || Boolean(step.current);
      }

      if (!currentGroup) continue;
      if (type === STEP_TYPE_PAUSE) {
        continue;
      }

      const subsetKey = String(
        step?.subsetId || step?.id || `${setName}-${loopIndex}-${loopTotal}`,
      );
      let subset = currentGroup.subsets.find((item) => item.key === subsetKey);
      if (!subset) {
        subset = {
          key: subsetKey,
          superset: Boolean(step?.superset),
          label: step?.subsetLabel,
          exercises: [],
        };
        currentGroup.subsets.push(subset);
      } else {
        subset.superset = subset.superset || Boolean(step?.superset);
        if (!subset.label && step?.subsetLabel) {
          subset.label = step.subsetLabel;
        }
      }

      const exercises = getExercises(step);
      if (exercises.length) {
        subset.exercises.push(...exercises);
      }
      currentGroup.hasSuperset = currentGroup.hasSuperset || subset.superset;
    }
    return groups;
  }, [remainingSteps]);

  // Next step name (just for the right panel)
  const nextStep = useMemo(() => {
    if (!session) return null;
    if (!session.running) {
      return session.steps[session.currentIndex];
    }
    let idx = session.currentIndex;
    while (idx < session.steps.length && session.steps[idx].completed) {
      idx += 1;
    }
    idx += 1;
    while (idx < session.steps.length && session.steps[idx].completed) {
      idx += 1;
    }
    return idx < session.steps.length ? session.steps[idx] : null;
  }, [session]);

  const nextSubsetStep = useMemo(() => {
    if (!session || !session.steps.length || !currentStep?.subsetId) {
      return null;
    }
    const subsetId = currentStep.subsetId;
    const startIdx = Math.max(session.currentIndex + 1, 0);
    for (let idx = startIdx; idx < session.steps.length; idx += 1) {
      const candidate = session.steps[idx];
      if (!candidate) continue;
      if (!candidate.subsetId) continue;
      if (candidate.subsetId !== subsetId) {
        return candidate;
      }
    }
    return null;
  }, [session, currentStep?.subsetId]);

  const nextStepExerciseLabels = useMemo(
    () => extractExerciseLabels(nextStep),
    [nextStep, extractExerciseLabels],
  );

  const hasFollowingSubsetExercises = useMemo(() => {
    if (!session || !nextStep?.subsetId) return false;
    const subsetId = nextStep.subsetId;
    return session.steps.some(
      (step, idx) => idx > session.currentIndex && step.subsetId === subsetId,
    );
  }, [session, nextStep]);

  const shouldShowNextExercises =
    Boolean(
      currentStep?.subsetId &&
        nextStep?.subsetId &&
        currentStep.subsetId === nextStep.subsetId,
    ) &&
    nextStepExerciseLabels.length > 0 &&
    hasFollowingSubsetExercises;

  const nextNameStep = shouldShowNextExercises
    ? nextStep
    : (nextSubsetStep ?? nextStep);

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
            {currentStep ? (
              currentStepPills.length ? (
                <div className="exercise-pills compact current-step-pills">
                  {currentStepPills.map((text, idx) => (
                    <span key={idx} className="pill">
                      {text}
                    </span>
                  ))}
                </div>
              ) : (
                currentExerciseLabel
              )
            ) : (
              "No training"
            )}
          </div>
        </div>

        <div className="next-card">
          <div className="label muted">Next</div>
          <div className="next-name">
            {nextNameStep ? (
              shouldShowNextExercises ? (
                <div className="exercise-pills compact next-step-pills">
                  {nextStepExerciseLabels.map((text, idx) => (
                    <span key={idx} className="pill">
                      {text}
                    </span>
                  ))}
                </div>
              ) : (
                getStepName(nextNameStep)
              )
            ) : (
              "None"
            )}
          </div>

          <div className="actions vertical session-actions">
            <button
              ref={runButtonRef}
              className="btn primary"
              onClick={running ? onPause : onStart}
              disabled={!session || done}
            >
              {startLabel}
            </button>
            <button
              ref={nextButtonRef}
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

      <div className="session-steps-cards">
        {!session ? <p className="muted">Start training to see sets.</p> : null}

        {groupedSteps.map((group) => {
          const roundText =
            group.loopTotal > 1
              ? ` • round ${group.loopIndex || 1}/${group.loopTotal}`
              : "";
          const pauseDuration =
            group.type === STEP_TYPE_PAUSE && group.estimatedSeconds
              ? ` • ${formatMillis(group.estimatedSeconds * 1000)}`
              : "";
          const groupTitle =
            group.setName && group.setName.trim()
              ? group.setName
              : group.type === STEP_TYPE_PAUSE
                ? "Pause"
                : "Step";
          const classes = [
            "set-card",
            group.hasSuperset ? "is-superset" : "is-normal",
            group.current ? "current" : "",
            group.type === STEP_TYPE_PAUSE ? "pause" : "",
          ]
            .filter(Boolean)
            .join(" ");

          return (
            <div key={group.key} className={classes}>
              <div className="set-header">
                <strong>
                  {groupTitle}
                  {group.type === STEP_TYPE_PAUSE ? pauseDuration : roundText}
                </strong>
              </div>

              {group.subsets.length ? (
                group.subsets.map((subset) => {
                  const showSubsetLabel =
                    subset.label && group.subsets.length > 1;
                  const pillKey = `${group.key}-${subset.key}`;
                  const applySuperset =
                    subset.superset || group.subsets.length > 1;
                  return (
                    <div
                      key={pillKey}
                      className={[
                        "exercise-pills-container",
                        applySuperset ? "is-superset" : "is-normal",
                      ]
                        .filter(Boolean)
                        .join(" ")}
                    >
                      {showSubsetLabel && (
                        <div className="subset-label muted small">
                          {subset.label}
                        </div>
                      )}
                      {subset.superset ? (
                        <span className="set-badge superset">Superset</span>
                      ) : null}
                      <div className="exercise-pills compact">
                        {subset.exercises.map((ex, idx) => {
                          const text = formatExerciseLine(ex);
                          if (!text) return null;
                          return (
                            <span key={`${pillKey}-${idx}`} className="pill">
                              {text}
                            </span>
                          );
                        })}
                      </div>
                    </div>
                  );
                })
              ) : group.type === STEP_TYPE_PAUSE ? null : (
                <div className="muted small">No exercises yet.</div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
