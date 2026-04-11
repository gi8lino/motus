import { useCallback, useMemo, type RefObject } from "react";
import FitnessCenterRoundedIcon from "@mui/icons-material/FitnessCenterRounded";
import PauseCircleRoundedIcon from "@mui/icons-material/PauseCircleRounded";
import PlayArrowRoundedIcon from "@mui/icons-material/PlayArrowRounded";
import SkipNextRoundedIcon from "@mui/icons-material/SkipNextRounded";
import TaskAltRoundedIcon from "@mui/icons-material/TaskAltRounded";
import TimerRoundedIcon from "@mui/icons-material/TimerRounded";
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Divider,
  Stack,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import { alpha } from "@mui/material/styles";

import { formatElapsedMillis, formatCountdownMillis } from "../../utils/format";
import { getCountdownDisplayMillis } from "../../utils/countdown";
import { PROMPTS } from "../../utils/messages";
import { UI_TEXT } from "../../utils/uiText";
import { STEP_TYPE_PAUSE } from "../../utils/step";
import type { Exercise, TrainingState, TrainingStepState } from "../../types";
import {
  buildExercisePills,
  buildStepGroups,
  extractExerciseLabels,
  formatStepCounter,
  getExercises,
  getNextStep,
  getNextSubsetStep,
  getStepName,
  hasFollowingSubsetExercises,
} from "../../utils/training";

function formatExerciseMetric(exercise: Exercise | undefined) {
  if (!exercise) return "";
  if (exercise.reps && exercise.weight) return `${exercise.reps} reps · ${exercise.weight}`;
  if (exercise.reps) return `${exercise.reps} reps`;
  if (exercise.duration && exercise.weight) return `${exercise.duration} · ${exercise.weight}`;
  if (exercise.duration) return exercise.duration;
  if (exercise.weight) return exercise.weight;
  return "";
}

function formatRoundLabel(step: TrainingStepState | null) {
  const total = step?.loopTotal ?? 0;
  if (total <= 1) return "";
  return `Round ${step?.loopIndex || 1}/${total}`;
}

function TrainingControls({
  training,
  done,
  running,
  startLabel,
  isLastStep,
  mobile,
  onStart,
  onPause,
  onNext,
  onFinish,
  runButtonRef,
  nextButtonRef,
}: {
  training: TrainingState | null;
  done: boolean | undefined;
  running: boolean | undefined;
  startLabel: string;
  isLastStep: boolean;
  mobile: boolean;
  onStart: () => void;
  onPause: () => void;
  onNext: () => void;
  onFinish: () => void;
  runButtonRef?: RefObject<HTMLButtonElement>;
  nextButtonRef?: RefObject<HTMLButtonElement>;
}) {
  const handleNext = () => {
    if (isLastStep) {
      onFinish();
      return;
    }
    onNext();
  };

  const runIcon = running ? <PauseCircleRoundedIcon /> : <PlayArrowRoundedIcon />;
  const nextIcon = isLastStep ? <TaskAltRoundedIcon /> : <SkipNextRoundedIcon />;

  return (
    <Stack
      spacing={1.25}
      sx={{
        width: "100%",
        p: mobile ? 1.25 : 0,
      }}
    >
      <Button
        ref={runButtonRef}
        variant="contained"
        size="large"
        fullWidth
        startIcon={runIcon}
        onClick={running ? onPause : onStart}
        disabled={!training || done}
        sx={{
          minHeight: mobile ? 76 : 58,
          fontSize: mobile ? "1.05rem" : "1rem",
        }}
      >
        {startLabel}
      </Button>

      <Button
        ref={nextButtonRef}
        variant="contained"
        color="secondary"
        size="large"
        fullWidth
        startIcon={nextIcon}
        onClick={handleNext}
        disabled={!training || done || !training.startedAt}
        sx={{
          minHeight: mobile ? 68 : 56,
          fontSize: mobile ? "1rem" : "0.98rem",
          opacity: mobile ? 0.96 : 1,
        }}
      >
        {isLastStep
          ? UI_TEXT.training.nextLabels.finish
          : UI_TEXT.training.nextLabels.next}
      </Button>
    </Stack>
  );
}

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
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down("md"));
  const running = training?.running;
  const done = training?.done;

  const isLastStep =
    training && training.steps.length
      ? training.currentIndex >= training.steps.length - 1
      : false;

  const hasProgress = training?.steps?.some(
    (step) => (step.elapsedMillis || 0) > 0 || Boolean(step.completed),
  );
  const hasStarted = Boolean(training?.running) || Boolean(hasProgress);

  const isAutoAdvance =
    (currentStep?.type === STEP_TYPE_PAUSE &&
      Boolean(currentStep.pauseOptions?.autoAdvance)) ||
    Boolean(currentStep?.autoAdvance);

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
    if (isAutoAdvance && currentStep.estimatedSeconds) {
      const durationMs = currentStep.estimatedSeconds * 1000;
      return getCountdownDisplayMillis(durationMs, elapsed);
    }
    return elapsed;
  }, [currentStep, elapsed, isAutoAdvance]);

  const clockText = useMemo(() => {
    if (!currentStep) return formatElapsedMillis(0, { showHours });
    return isAutoAdvance
      ? formatCountdownMillis(displayMillis, { showHours })
      : formatElapsedMillis(displayMillis, { showHours });
  }, [currentStep, displayMillis, isAutoAdvance, showHours]);

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
    () => (training?.steps || []).filter((step) => !step.completed),
    [training?.trainingId, training?.steps],
  );

  const groupedSteps = useMemo(
    () => buildStepGroups(remainingSteps),
    [remainingSteps],
  );

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

  const activeExercises = useMemo(() => getExercises(currentStep), [currentStep]);
  const primaryExercise = activeExercises[0];
  const heroTitle = currentStep
    ? primaryExercise?.name || getStepName(currentStep)
    : PROMPTS.noTraining;
  const heroMetric =
    currentStep?.type === STEP_TYPE_PAUSE
      ? currentStep.estimatedSeconds
        ? formatCountdownMillis(currentStep.estimatedSeconds * 1000, {
            showHours,
          })
        : UI_TEXT.labels.pause
      : formatExerciseMetric(primaryExercise);
  const heroChips =
    currentStep?.type === STEP_TYPE_PAUSE
      ? []
      : currentStepPills.length > 1
        ? currentStepPills.slice(1)
        : [];
  const roundLabel = formatRoundLabel(currentStep);
  const currentTone =
    currentStep?.type === STEP_TYPE_PAUSE ? "pause" : training ? "live" : "idle";

  const nextPrimaryLabel = shouldShowNextExercises
    ? nextStepExerciseLabels[0]
    : nextNameStep
      ? getStepName(nextNameStep)
      : UI_TEXT.training.nextLabels.none;
  const nextSecondaryLabels = shouldShowNextExercises
    ? nextStepExerciseLabels.slice(1)
    : nextNameStep && nextStepExerciseLabels.length
      ? nextStepExerciseLabels
      : [];

  return (
    <Stack spacing={2.25}>
      <Box
        sx={{
          display: "grid",
          gap: 2.25,
          gridTemplateColumns: { xs: "1fr", lg: "minmax(0, 1.7fr) 340px" },
          alignItems: "start",
        }}
      >
        <Card
          sx={{
            border: 1,
            borderColor:
              currentTone === "pause" ? "warning.main" : "primary.main",
            background:
              currentTone === "pause"
                ? `linear-gradient(160deg, ${alpha(theme.palette.warning.main, 0.18)}, ${alpha(theme.palette.background.paper, 0.96)})`
                : `linear-gradient(160deg, ${alpha(theme.palette.primary.main, 0.16)}, ${alpha(theme.palette.background.paper, 0.96)})`,
          }}
        >
          <CardContent sx={{ p: { xs: 2, md: 3 } }}>
            <Stack spacing={2.5}>
              <Stack
                direction={{ xs: "column", sm: "row" }}
                justifyContent="space-between"
                alignItems={{ xs: "flex-start", sm: "center" }}
                spacing={1}
              >
                <Stack
                  direction="row"
                  spacing={1}
                  alignItems="center"
                  useFlexGap
                  flexWrap="wrap"
                >
                  <Chip
                    size="small"
                    icon={
                      currentTone === "pause" ? (
                        <PauseCircleRoundedIcon />
                      ) : (
                        <FitnessCenterRoundedIcon />
                      )
                    }
                    label={UI_TEXT.training.cards.now}
                    color={currentTone === "pause" ? "warning" : "primary"}
                  />
                  {training ? (
                    <Chip
                      size="small"
                      variant="outlined"
                      label={formatStepCounter(currentNumber, totalSteps)}
                    />
                  ) : null}
                  {roundLabel ? <Chip size="small" label={roundLabel} /> : null}
                </Stack>

                {workoutName ? (
                  <Typography variant="body2" color="text.secondary">
                    {workoutName}
                  </Typography>
                ) : null}
              </Stack>

              <Stack spacing={1.5}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <TimerRoundedIcon color="action" />
                  <Typography
                    variant={isMobile ? "h2" : "h1"}
                    sx={{
                      fontVariantNumeric: "tabular-nums",
                      letterSpacing: "-0.06em",
                      lineHeight: 1,
                    }}
                  >
                    {clockText}
                  </Typography>
                </Stack>

                <Box>
                  <Typography
                    variant={isMobile ? "h4" : "h2"}
                    sx={{
                      textTransform: currentStep ? "uppercase" : "none",
                      letterSpacing: currentStep ? "0.02em" : undefined,
                    }}
                  >
                    {heroTitle}
                  </Typography>

                  {heroMetric ? (
                    <Typography
                      variant={isMobile ? "h6" : "h5"}
                      color="text.secondary"
                      sx={{ mt: 0.75 }}
                    >
                      {heroMetric}
                    </Typography>
                  ) : null}
                </Box>
              </Stack>

              {heroChips.length ? (
                <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                  {heroChips.map((text, idx) => (
                    <Chip key={`${text}-${idx}`} label={text} />
                  ))}
                </Stack>
              ) : null}

              {!training ? (
                <Typography color="text.secondary">
                  {UI_TEXT.training.states.startToSeeSets}
                </Typography>
              ) : null}
            </Stack>
          </CardContent>
        </Card>

        <Stack spacing={2.25}>
          <Card sx={{ border: 1, borderColor: "divider" }}>
            <CardContent sx={{ p: 2.25 }}>
              <Stack spacing={1.5}>
                <Stack
                  direction="row"
                  justifyContent="space-between"
                  alignItems="center"
                >
                  <Typography variant="overline" color="text.secondary">
                    {UI_TEXT.training.cards.next}
                  </Typography>
                  {nextNameStep?.estimatedSeconds ? (
                    <Chip
                      size="small"
                      variant="outlined"
                      label={formatCountdownMillis(
                        nextNameStep.estimatedSeconds * 1000,
                        { showHours },
                      )}
                    />
                  ) : null}
                </Stack>

                <Typography variant="h5">{nextPrimaryLabel}</Typography>

                {nextSecondaryLabels.length ? (
                  <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                    {nextSecondaryLabels.map((text, idx) => (
                      <Chip key={`${text}-${idx}`} size="small" label={text} />
                    ))}
                  </Stack>
                ) : null}

                {!nextNameStep ? (
                  <Typography variant="body2" color="text.secondary">
                    Finish is the next move.
                  </Typography>
                ) : null}
              </Stack>
            </CardContent>
          </Card>

          {!isMobile ? (
            <Card sx={{ border: 1, borderColor: "divider" }}>
              <CardContent sx={{ p: 2 }}>
                <Typography
                  variant="overline"
                  color="text.secondary"
                  sx={{ mb: 1.25, display: "block" }}
                >
                  Controls
                </Typography>
                <TrainingControls
                  training={training}
                  done={done}
                  running={running}
                  startLabel={startLabel}
                  isLastStep={isLastStep}
                  mobile={false}
                  onStart={onStart}
                  onPause={onPause}
                  onNext={handleNext}
                  onFinish={onFinish}
                  runButtonRef={runButtonRef}
                  nextButtonRef={nextButtonRef}
                />
              </CardContent>
            </Card>
          ) : null}
        </Stack>
      </Box>

      <Card sx={{ border: 1, borderColor: "divider" }}>
        <CardContent sx={{ p: { xs: 2, md: 2.5 } }}>
          <Stack spacing={1.75}>
            <Box>
              <Typography variant="h6">Queue</Typography>
              <Typography variant="body2" color="text.secondary">
                Upcoming work and recovery blocks.
              </Typography>
            </Box>

            {!training ? (
              <Typography color="text.secondary">
                {UI_TEXT.training.states.startToSeeSets}
              </Typography>
            ) : null}
            {training && !groupedSteps.length ? (
              <Typography color="text.secondary">
                Training wrapped. Finish when you're ready.
              </Typography>
            ) : null}

            <Stack spacing={1.25}>
              {groupedSteps.map((group) => {
                const roundText =
                  group.loopTotal > 1
                    ? `Round ${group.loopIndex || 1}/${group.loopTotal}`
                    : "";
                const pauseDuration =
                  group.type === STEP_TYPE_PAUSE && group.estimatedSeconds
                    ? formatCountdownMillis(group.estimatedSeconds * 1000, {
                        showHours,
                      })
                    : "";
                const groupTitle =
                  group.setName && group.setName.trim()
                    ? group.setName
                    : group.type === STEP_TYPE_PAUSE
                      ? UI_TEXT.labels.pause
                      : UI_TEXT.labels.step;

                return (
                  <Card
                    key={group.key}
                    variant="outlined"
                    sx={{
                      borderColor: group.current
                        ? "primary.main"
                        : "divider",
                      bgcolor: group.current
                        ? alpha(theme.palette.primary.main, 0.08)
                        : "transparent",
                    }}
                  >
                    <CardContent sx={{ p: 2 }}>
                      <Stack spacing={1.25}>
                        <Stack
                          direction={{ xs: "column", sm: "row" }}
                          spacing={1}
                          justifyContent="space-between"
                          alignItems={{ xs: "flex-start", sm: "center" }}
                        >
                          <Stack
                            direction="row"
                            spacing={1}
                            useFlexGap
                            flexWrap="wrap"
                          >
                            <Typography variant="subtitle1" fontWeight={700}>
                              {groupTitle}
                            </Typography>
                            {group.current ? (
                              <Chip
                                size="small"
                                color="primary"
                                label="Current"
                              />
                            ) : null}
                            {group.hasSuperset ? (
                              <Chip size="small" label="Superset" />
                            ) : null}
                          </Stack>

                          <Stack direction="row" spacing={1}>
                            {roundText ? (
                              <Chip size="small" variant="outlined" label={roundText} />
                            ) : null}
                            {pauseDuration ? (
                              <Chip
                                size="small"
                                color="warning"
                                variant="outlined"
                                label={pauseDuration}
                              />
                            ) : null}
                          </Stack>
                        </Stack>

                        {group.subsets.length ? (
                          <Stack spacing={1.25}>
                            {group.subsets.map((subset, subsetIndex) => (
                              <Box key={`${group.key}-${subset.key}`}>
                                {subsetIndex > 0 ? <Divider sx={{ mb: 1.25 }} /> : null}
                                {subset.label && group.subsets.length > 1 ? (
                                  <Typography
                                    variant="caption"
                                    color="text.secondary"
                                    sx={{
                                      textTransform: "uppercase",
                                      letterSpacing: "0.12em",
                                      display: "block",
                                      mb: 0.75,
                                    }}
                                  >
                                    {subset.label}
                                  </Typography>
                                ) : null}

                                <Stack
                                  direction="row"
                                  spacing={1}
                                  useFlexGap
                                  flexWrap="wrap"
                                >
                                  {buildExercisePills(subset.exercises).map(
                                    (text, idx) => (
                                      <Chip
                                        key={`${subset.key}-${idx}`}
                                        size="small"
                                        label={text}
                                      />
                                    ),
                                  )}
                                </Stack>
                              </Box>
                            ))}
                          </Stack>
                        ) : group.type === STEP_TYPE_PAUSE ? null : (
                          <Typography variant="body2" color="text.secondary">
                            {UI_TEXT.training.states.noExercises}
                          </Typography>
                        )}
                      </Stack>
                    </CardContent>
                  </Card>
                );
              })}
            </Stack>
          </Stack>
        </CardContent>
      </Card>

      {isMobile ? (
        <Box
          sx={{
            position: "sticky",
            bottom: 16,
            zIndex: 9,
          }}
        >
          <Card
            sx={{
              border: 1,
              borderColor: alpha(theme.palette.primary.main, 0.35),
              bgcolor: alpha(theme.palette.background.paper, 0.95),
              backdropFilter: "blur(12px)",
            }}
          >
            <TrainingControls
              training={training}
              done={done}
              running={running}
              startLabel={startLabel}
              isLastStep={isLastStep}
              mobile
              onStart={onStart}
              onPause={onPause}
              onNext={handleNext}
              onFinish={onFinish}
              runButtonRef={runButtonRef}
              nextButtonRef={nextButtonRef}
            />
          </Card>
        </Box>
      ) : null}
    </Stack>
  );
}
