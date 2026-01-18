import { useCallback, useMemo, type RefObject } from "react";
import { formatElapsedMillis, formatCountdownMillis } from "../../utils/format";
import { getCountdownDisplayMillis } from "../../utils/countdown";
import { PROMPTS } from "../../utils/messages";
import { UI_TEXT } from "../../utils/uiText";
import { STEP_TYPE_PAUSE } from "../../utils/step";
import type { TrainingState, TrainingStepState } from "../../types";
import {
  buildExercisePills,
  buildStepGroups,
  extractExerciseLabels,
  getCurrentExerciseLabel,
  getNextStep,
  getNextSubsetStep,
  getStepName,
  hasFollowingSubsetExercises,
  formatStepCounter,
} from "../../utils/training";

// TrainingCard renders the main training status card and controls.
// NOTE: This component is intentionally "dumb":
// - no timers
// - no scheduling
// - no logging
export function TrainingCard({
  training,
  currentStep,
  elapsed,
  workoutName,
  showHours,
  onStart,
  onPause,
  onNext,
  onFinish,
  onStopAudio,
  runButtonRef,
  nextButtonRef,
}: {
  training: TrainingState | null;
  currentStep: TrainingStepState | null;
  elapsed: number;
  workoutName?: string;
  showHours?: boolean;
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
    ? UI_TEXT.actions.pause
    : hasStarted
      ? UI_TEXT.actions.continue
      : UI_TEXT.actions.start;

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
      return getCountdownDisplayMillis(durationMs, elapsed);
    }

    // Otherwise show elapsed.
    return elapsed;
  }, [currentStep, elapsed, isAutoAdvance]);

  const clockText = useMemo(() => {
    if (!currentStep) return formatElapsedMillis(0, { showHours });
    return isAutoAdvance
      ? formatCountdownMillis(displayMillis, { showHours })
      : formatElapsedMillis(displayMillis, { showHours });
  }, [currentStep, displayMillis, isAutoAdvance, showHours]);

  const currentExerciseLabel = useMemo(() => {
    if (!currentStep) return PROMPTS.noTraining;
    return getCurrentExerciseLabel(currentStep as any);
  }, [currentStep]);

  const extractLabels = useCallback(
    (step: TrainingStepState | null, startIndex = 0) =>
      extractExerciseLabels(step, training, startIndex),
    [training],
  );

  const currentStepPills = useMemo(
    () => extractLabels(currentStep),
    [currentStep, extractLabels],
  );

  const totalSteps = training?.steps?.length || 0;
  const currentNumber = training ? training.currentIndex + 1 : 0;

  const remainingSteps = useMemo(
    () => (training?.steps || []).filter((s: any) => !s.completed),
    [training?.trainingId, training?.steps],
  );

  const groupedSteps = useMemo(
    () => buildStepGroups(remainingSteps as TrainingStepState[]),
    [remainingSteps],
  );

  // Next step name (just for the right panel)
  const nextStep = useMemo(() => getNextStep(training), [training]);

  const nextSubsetStep = useMemo(
    () => getNextSubsetStep(training, currentStep),
    [training, currentStep],
  );

  const nextStepExerciseLabels = useMemo(
    () => extractLabels(nextStep),
    [nextStep, extractLabels],
  );

  const hasFollowingSubset = useMemo(
    () => hasFollowingSubsetExercises(training, nextStep),
    [training, nextStep],
  );

  const shouldShowNextExercises =
    Boolean(
      currentStep?.subsetId &&
        nextStep?.subsetId &&
        currentStep.subsetId === nextStep.subsetId,
    ) &&
    nextStepExerciseLabels.length > 0 &&
    hasFollowingSubset;

  const nextNameStep = shouldShowNextExercises
    ? nextStep
    : (nextSubsetStep ?? nextStep);

  return (
    <div className="training-card">
      <div className="training-main">
        <div className="current-card">
          <div className="label muted">{UI_TEXT.training.cards.now}</div>

          {workoutName ? (
            <div className="muted small">{workoutName}</div>
          ) : null}

          {training ? (
            <div className="muted small">
              {formatStepCounter(currentNumber, totalSteps)}
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
              PROMPTS.noTraining
            )}
          </div>
        </div>

        <div className="next-card">
          <div className="label muted">{UI_TEXT.training.cards.next}</div>
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
              UI_TEXT.training.nextLabels.none
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
              {isLastStep
                ? UI_TEXT.training.nextLabels.finish
                : UI_TEXT.training.nextLabels.next}
            </button>
          </div>
        </div>
      </div>

      <div className="training-steps-cards">
        {!training ? (
          <p className="muted">{UI_TEXT.training.states.startToSeeSets}</p>
        ) : null}

        {groupedSteps.map((group) => {
          const roundText =
            group.loopTotal > 1
              ? ` • round ${group.loopIndex || 1}/${group.loopTotal}`
              : "";
          const pauseDuration =
            group.type === STEP_TYPE_PAUSE && group.estimatedSeconds
              ? ` • ${formatCountdownMillis(group.estimatedSeconds * 1000, {
                  showHours,
                })}`
              : "";
          const groupTitle =
            group.setName && group.setName.trim()
              ? group.setName
              : group.type === STEP_TYPE_PAUSE
                ? UI_TEXT.labels.pause
                : UI_TEXT.labels.step;

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
                        {buildExercisePills(subset.exercises).map(
                          (text, idx) => (
                            <span key={`${pillKey}-${idx}`} className="pill">
                              {text}
                            </span>
                          ),
                        )}
                      </div>
                    </div>
                  );
                })
              ) : group.type === STEP_TYPE_PAUSE ? null : (
                <div className="muted small">
                  {UI_TEXT.training.states.noExercises}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
