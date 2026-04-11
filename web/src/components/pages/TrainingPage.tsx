import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import PlayCircleFilledRoundedIcon from "@mui/icons-material/PlayCircleFilledRounded";
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  Typography,
} from "@mui/material";

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
import { TrainingFinishModal } from "../training/FinishTrainingModal";
import { TrainingOverrunModal } from "../training/OverrunTrainingModal";
import { useTrainingAudio } from "../../hooks/useTrainingAudio";
import { useTrainingOverrun } from "../../hooks/useTrainingOverrun";
import { useTrainingKeyboard } from "../../hooks/useTrainingKeyboard";

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
  showHours: boolean;
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

  const selectedWorkout = useMemo(
    () => workouts.find((workout) => workout.id === selectedWorkoutId) ?? null,
    [selectedWorkoutId, workouts],
  );

  return (
    <>
      <Card
        sx={{
          overflow: "visible",
          border: 1,
          borderColor: "divider",
          bgcolor: "background.paper",
        }}
      >
        <CardContent sx={{ p: { xs: 2, md: 3 } }}>
          <Stack spacing={2.5}>
            <Stack
              direction={{ xs: "column", md: "row" }}
              justifyContent="space-between"
              spacing={2}
            >
              <Box>
                <Typography variant="h4" sx={{ mb: 0.75 }}>
                  {UI_TEXT.pages.training.title}
                </Typography>

                {headerStatus ? (
                  <Stack
                    direction="row"
                    spacing={1}
                    useFlexGap
                    flexWrap="wrap"
                    sx={{ alignItems: "center" }}
                  >
                    <Chip
                      size="small"
                      color="primary"
                      label={
                        workoutName || UI_TEXT.training.headers.workoutFallback
                      }
                    />
                    <Chip size="small" label={headerStatus} />
                    <Chip
                      size="small"
                      variant="outlined"
                      label={formatElapsedMillis(elapsed, {
                        showHours: data.showHours,
                      })}
                    />
                  </Stack>
                ) : (
                  <Typography color="text.secondary">
                    {PROMPTS.selectWorkoutToStart}
                  </Typography>
                )}
              </Box>

              <Stack
                direction={{ xs: "column", sm: "row" }}
                spacing={1.25}
                sx={{
                  width: { xs: "100%", md: "auto" },
                  alignItems: { xs: "stretch", md: "center" },
                }}
              >
                <FormControl
                  size="small"
                  sx={{ minWidth: { xs: "100%", sm: 280 } }}
                >
                  <InputLabel id="training-workout-label">
                    {UI_TEXT.labels.workout}
                  </InputLabel>
                  <Select
                    labelId="training-workout-label"
                    value={selectedWorkoutId ?? ""}
                    label={UI_TEXT.labels.workout}
                    onChange={(event) =>
                      onSelectWorkout(String(event.target.value))
                    }
                  >
                    <MenuItem value="">
                      <em>{UI_TEXT.placeholders.selectWorkout}</em>
                    </MenuItem>
                    {workouts.map((workout) => (
                      <MenuItem key={workout.id} value={workout.id}>
                        {workout.name} · {workout.steps.length} steps
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>

                <Button
                  variant="contained"
                  size="large"
                  startIcon={<PlayCircleFilledRoundedIcon />}
                  onClick={onStartTraining}
                  disabled={startDisabled}
                  title={startTitle}
                  sx={{
                    minWidth: { xs: "100%", sm: 148 },
                    alignSelf: "stretch",
                  }}
                >
                  {startLabel}
                </Button>
              </Stack>
            </Stack>

            {selectedWorkout && !training ? (
              <Box
                sx={{
                  px: 1.5,
                  py: 1.25,
                  borderRadius: 3,
                  bgcolor: "action.hover",
                }}
              >
                <Typography variant="body2" color="text.secondary">
                  {selectedWorkout.name} is ready with {selectedWorkout.steps.length}{" "}
                  steps. Hit start to jump into the first exercise.
                </Typography>
              </Box>
            ) : null}

            <TrainingCard
              training={training}
              currentStep={currentStep}
              elapsed={elapsed}
              workoutName={workoutName}
              showHours={data.showHours}
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
          </Stack>
        </CardContent>
      </Card>

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
        showHours={data.showHours}
      />
    </>
  );
}
