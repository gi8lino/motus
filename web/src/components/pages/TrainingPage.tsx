import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import type {
  TrainingState,
  TrainingStepState,
  SoundOption,
  Workout,
} from "../../types";
import { formatElapsedMillis } from "../../utils/format";
import { PROMPTS } from "../../utils/messages";
import { UI_TEXT } from "../../utils/uiText";
import { getTrainingHeaderStatus } from "../../utils/training";
import { TrainingCard } from "../training/TrainingCard";
import { WorkoutPicker } from "../workouts/WorkoutPicker";
import { TrainingFinishModal } from "../training/FinishTrainingModal";
import { TrainingOverrunModal } from "../training/OverrunTrainingModal";
import { useTrainingAudio } from "../../hooks/useTrainingAudio";
import { useTrainingOverrun } from "../../hooks/useTrainingOverrun";
import { useTrainingKeyboard } from "../../hooks/useTrainingKeyboard";

// Training runs the active workout training.
export type TrainingViewData = {
  workouts: Workout[];
  selectedWorkoutId: string | null;
  startDisabled: boolean;
  startTitle?: string;
  training: TrainingState | null;
  currentStep: TrainingStepState | null;
  elapsed: number;
  workoutName: string;
  sounds: SoundOption[];
  pauseOnTabHidden: boolean;
};

export type TrainingViewActions = {
  onSelectWorkout: (id: string) => void;
  onStartTraining: () => void | Promise<void>;
  markSoundPlayed: () => void;
  onStartStep: () => void;
  onPause: () => void;
  onNext: () => void;
  onFinishTraining: () => Promise<string | null>;
  onCopySummary: () => void;
  onToast: (message: string) => void;
};

export function TrainingView({
  data,
  actions,
}: {
  data: TrainingViewData;
  actions: TrainingViewActions;
}) {
  const {
    workouts,
    selectedWorkoutId,
    startDisabled,
    startTitle,
    training,
    currentStep,
    elapsed,
    workoutName,
    sounds,
    pauseOnTabHidden,
  } = data;
  const {
    onSelectWorkout,
    onStartTraining,
    markSoundPlayed,
    onStartStep,
    onPause,
    onNext,
    onFinishTraining,
    onCopySummary,
    onToast,
  } = actions;
  const [finishSummary, setFinishSummary] = useState<string | null>(null);
  const autoFinishRef = useRef<string | null>(null);

  // ---------- Refs for stable handlers ----------
  const trainingRef = useRef<TrainingState | null>(training);
  const currentStepRef = useRef<TrainingStepState | null>(currentStep);
  const elapsedRef = useRef(elapsed);
  useEffect(() => {
    trainingRef.current = training;
  }, [training]);
  useEffect(() => {
    currentStepRef.current = currentStep;
  }, [currentStep]);
  useEffect(() => {
    elapsedRef.current = elapsed;
  }, [elapsed]);

  const { handlePause, handleStart, stopActiveAudio } = useTrainingAudio({
    training,
    currentStep,
    sounds,
    markSoundPlayed,
    onPause,
    onStartStep,
    onToast,
    pauseOnTabHidden,
    refs: { trainingRef, currentStepRef, elapsedRef },
  });

  const {
    overrunModal,
    overrunCountdown,
    overrunModalRef,
    handleOverrunPause,
    handleOverrunPostpone,
  } = useTrainingOverrun({
    training,
    currentStep,
    elapsed,
    onPause: handlePause,
    refs: { trainingRef, currentStepRef, elapsedRef },
  });

  // ---------- Keyboard shortcuts (ignore key repeat) ----------
  const runButtonRef = useRef<HTMLButtonElement>(null!);
  const nextActionButtonRef = useRef<HTMLButtonElement>(null!);

  useTrainingKeyboard({
    trainingRef,
    overrunModalRef,
    handleOverrunPostpone,
    handleOverrunPause,
    handlePause,
    handleStart,
    stopActiveAudio,
    onNext,
    onFinishTraining,
    setFinishSummary,
  });

  const handleFinish = useCallback(async () => {
    const summary = await onFinishTraining();
    if (summary) setFinishSummary(summary);
  }, [onFinishTraining]);

  useEffect(() => {
    if (!training?.done) return;
    if (finishSummary) return;
    if (autoFinishRef.current === training.trainingId) return;
    autoFinishRef.current = training.trainingId;
    onFinishTraining().then((summary) => {
      if (summary) setFinishSummary(summary);
    });
  }, [training, finishSummary, onFinishTraining]);

  const headerStatus = useMemo(
    () => getTrainingHeaderStatus(training),
    [training],
  );

  const startLabel =
    selectedWorkoutId && training?.workoutId === selectedWorkoutId
      ? UI_TEXT.actions.new
      : UI_TEXT.actions.select;

  return (
    <>
      <section className="panel">
        <div className="panel-header">
          <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
            <h3 style={{ margin: 0 }}>{UI_TEXT.pages.training.title}</h3>
            <div className="muted small">
              {headerStatus ? (
                <>
                  <strong>
                    {workoutName || UI_TEXT.training.headers.workoutFallback}
                  </strong>
                  {" • "}
                  {headerStatus}
                  {" • "}
                  {formatElapsedMillis(elapsed)}
                </>
              ) : (
                <span>{PROMPTS.selectWorkoutToStart}</span>
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
              onClick={onStartTraining}
              disabled={startDisabled}
              title={startTitle}
            >
              {startLabel}
            </button>
          </div>
        </div>

        <TrainingCard
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
