import { useCallback, useMemo, type RefObject } from "react";
import { formatExerciseLine, formatMillis } from "../../utils/format";
import { STEP_TYPE_PAUSE } from "../../utils/step";
import type { Exercise, TrainngState, TrainngStepState } from "../../types";

type AnyStep = any;

// formatCountdownMillis formats a remaining-time countdown so that
// 19.9s displays as "00:20" (ceil), matching user expectation.
function formatCountdownMillis(ms: number): string {
  if (!Number.isFinite(ms) || ms <= 0) return "00:00";
  const totalSeconds = Math.ceil(ms / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
}

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

// TrainCard renders the main training status card and controls.
// NOTE: This component is intentionally "dumb":
// - no timers
// - no scheduling
// - no logging
export function TrainCard({
  training,
  currentStep,
  elapsed,
  workoutName,
  onStart,
  onPause,
  onNext,
  onFinish,
  onStopAudio,
  runButtonRef,
  nextButtonRef,
}: {
  training: TrainngState | null;
  currentStep: TrainngStepState | null;
  elapsed: number;
  workoutName?: string;
  onStart: () => void;
  onPause: () => void;
  onNext: () => void;
  onFinish: () => void;
  onStopAudio?: () => void;
  runButtonRef?: RefObject<HTMLButtonElement>;
  nextButtonRef?: RefObject<HTMLButtonElement>;
}) {
  const running = training?.running;
  const done = training?.done;

  const isLastStep =
    training && training.steps.length
      ? training.currentIndex >= training.steps.length - 1
      : false;

  const hasProgress = training?.steps?.some(
    (s: any) => (s.elapsedMillis || 0) > 0 || Boolean(s.completed),
  );
  const hasStarted = Boolean(training?.running) || Boolean(hasProgress);

  // If the step is auto-advance (countdown/pause with autoAdvance) and has a target,
  // show remaining; otherwise show elapsed.
  const isAutoAdvance =
    (currentStep?.type === STEP_TYPE_PAUSE &&
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      Boolean((currentStep as any).pauseOptions?.autoAdvance)) ||
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    Boolean((currentStep as any)?.autoAdvance);

  const startLabel = training?.running
    ? "Pause"
    : hasStarted
      ? "Continue"
      : "Start";

  const handleNext = useCallback(() => {
    onStopAudio?.();
    if (isLastStep) {
      onFinish();
      return;
    }
    onNext();
  }, [isLastStep, onFinish, onNext, onStopAudio]);

  const displayMillis = useMemo(() => {
    if (!currentStep) return elapsed;

    // For auto-advance steps with a duration, show remaining time.
    if (isAutoAdvance && currentStep.estimatedSeconds) {
      const durationMs = currentStep.estimatedSeconds * 1000;
      return Math.max(0, durationMs - Math.max(0, elapsed));
    }

    // Otherwise show elapsed.
    return elapsed;
  }, [currentStep, elapsed, isAutoAdvance]);

  const clockText = useMemo(() => {
    if (!currentStep) return "00:00";
    return isAutoAdvance
      ? formatCountdownMillis(displayMillis)
      : formatMillis(displayMillis);
  }, [currentStep, displayMillis, isAutoAdvance]);

  const currentExerciseLabel = useMemo(() => {
    if (!currentStep) return "No training";
    return getCurrentExerciseLabel(currentStep as any);
  }, [currentStep]);

  const extractExerciseLabels = useCallback(
    (step: TrainngStepState | null, startIndex = 0) => {
      if (!step) return [];

      if (step.type === STEP_TYPE_PAUSE) {
        const durationText = step.estimatedSeconds
          ? formatMillis(step.estimatedSeconds * 1000)
          : "";
        const pauseText = durationText ? `Pause • ${durationText}` : "Pause";
        return [pauseText];
      }

      if (step.superset && step.subsetId && training?.steps?.length) {
        const subsetId = step.subsetId;
        const seen = new Set<string>();
        const labels: string[] = [];
        for (let idx = startIndex; idx < training.steps.length; idx += 1) {
          const candidate = training.steps[idx];
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
    [training?.steps],
  );

  const currentStepPills = useMemo(
    () => extractExerciseLabels(currentStep),
    [currentStep, extractExerciseLabels],
  );

  const totalSteps = training?.steps?.length || 0;
  const currentNumber = training ? training.currentIndex + 1 : 0;

  const remainingSteps = useMemo(
    () => (training?.steps || []).filter((s: any) => !s.completed),
    [training?.trainingId, training?.steps],
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
      if (type === STEP_TYPE_PAUSE) continue;

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
      if (exercises.length) subset.exercises.push(...exercises);
      currentGroup.hasSuperset = currentGroup.hasSuperset || subset.superset;
    }

    return groups;
  }, [remainingSteps]);

  // Next step name (just for the right panel)
  const nextStep = useMemo(() => {
    if (!training) return null;

    if (!training.running) {
      return training.steps[training.currentIndex];
    }

    let idx = training.currentIndex;
    while (idx < training.steps.length && training.steps[idx].completed)
      idx += 1;
    idx += 1;
    while (idx < training.steps.length && training.steps[idx].completed)
      idx += 1;

    return idx < training.steps.length ? training.steps[idx] : null;
  }, [training]);

  const nextSubsetStep = useMemo(() => {
    if (!training || !training.steps.length || !currentStep?.subsetId)
      return null;

    const subsetId = currentStep.subsetId;
    const startIdx = Math.max(training.currentIndex + 1, 0);

    for (let idx = startIdx; idx < training.steps.length; idx += 1) {
      const candidate = training.steps[idx];
      if (!candidate) continue;
      if (!candidate.subsetId) continue;
      if (candidate.subsetId !== subsetId) return candidate;
    }

    return null;
  }, [training, currentStep?.subsetId]);

  const nextStepExerciseLabels = useMemo(
    () => extractExerciseLabels(nextStep),
    [nextStep, extractExerciseLabels],
  );

  const hasFollowingSubsetExercises = useMemo(() => {
    if (!training || !nextStep?.subsetId) return false;
    const subsetId = nextStep.subsetId;
    return training.steps.some(
      (step, idx) => idx > training.currentIndex && step.subsetId === subsetId,
    );
  }, [training, nextStep]);

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
    <div className="training-card">
      <div className="training-main">
        <div className="current-card">
          <div className="label muted">Now</div>

          {workoutName ? (
            <div className="muted small">{workoutName}</div>
          ) : null}

          {training ? (
            <div className="muted small">
              Step {currentNumber}/{totalSteps}
            </div>
          ) : null}

          <div className="clock-row">
            <div className="clock">{clockText}</div>
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

          <div className="actions vertical training-actions">
            <button
              ref={runButtonRef}
              className="btn primary"
              onClick={running ? onPause : onStart}
              disabled={!training || done}
            >
              {startLabel}
            </button>

            <button
              ref={nextButtonRef}
              className="btn large next"
              onClick={handleNext}
              disabled={!training || done || !training.startedAt}
            >
              {isLastStep ? "Finish" : "Next"}
            </button>
          </div>
        </div>
      </div>

      <div className="training-steps-cards">
        {!training ? (
          <p className="muted">Start training to see sets.</p>
        ) : null}

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
                      {showSubsetLabel ? (
                        <div className="subset-label muted small">
                          {subset.label}
                        </div>
                      ) : null}

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
