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
  if (exercise.reps && exercise.weight) {
    return `${exercise.reps} reps · ${exercise.weight}`;
  }
  if (exercise.reps) return `${exercise.reps} reps`;
  if (exercise.duration && exercise.weight) {
    return `${exercise.duration} · ${exercise.weight}`;
  }
  if (exercise.duration) return exercise.duration;
  if (exercise.weight) return exercise.weight;
  return "";
}

function formatRoundLabel(step: TrainingStepState | null) {
  const total = step?.loopTotal ?? 0;
  if (total <= 1) return "";
  return `Round ${step?.loopIndex || 1}/${total}`;
}

function formatNextSupportLine(
  nextNameStep: TrainingStepState | null,
  nextSecondaryLabels: string[],
  showHours?: boolean,
) {
  if (!nextNameStep) return "Finish is the next move.";
  if (nextSecondaryLabels.length) return nextSecondaryLabels.join(" • ");
  if (nextNameStep.estimatedSeconds) {
    return `Target ${formatCountdownMillis(nextNameStep.estimatedSeconds * 1000, {
      showHours,
    })}`;
  }
  return "Ready when you are.";
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

  return (
    <Stack spacing={1.25} sx={{ width: "100%", p: mobile ? 1.25 : 0 }}>
      <Button
        ref={runButtonRef}
        variant="contained"
        size="large"
        fullWidth
        startIcon={
          running ? <PauseCircleRoundedIcon /> : <PlayArrowRoundedIcon />
        }
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
        startIcon={
          isLastStep ? <TaskAltRoundedIcon /> : <SkipNextRoundedIcon />
        }
        onClick={handleNext}
        disabled={!training || done || !training.startedAt}
        sx={{
          minHeight: mobile ? 68 : 56,
          fontSize: mobile ? "1rem" : "0.98rem",
          opacity: mobile ? 0.98 : 1,
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
  const heroSupport =
    currentStep?.type === STEP_TYPE_PAUSE
      ? "Recovery block"
      : currentStepPills.length > 1
        ? currentStepPills.slice(1).join(" • ")
        : workoutName || "Current movement";
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
  const nextSupportLine = formatNextSupportLine(
    nextNameStep,
    nextSecondaryLabels,
    showHours,
  );

  return (
    <Stack spacing={2.5}>
      <Box
        sx={{
          display: "grid",
          gap: 2.5,
          gridTemplateColumns: { xs: "1fr", lg: "minmax(0, 1.75fr) 320px" },
          alignItems: "start",
        }}
      >
        <Card
          sx={{
            background:
              currentTone === "pause"
                ? `linear-gradient(150deg, ${alpha(theme.palette.warning.main, 0.12)}, ${alpha(theme.palette.background.paper, 0.98)})`
                : `linear-gradient(150deg, ${alpha(theme.palette.primary.main, 0.1)}, ${alpha(theme.palette.background.paper, 0.98)})`,
          }}
        >
          <CardContent sx={{ p: { xs: 2.25, md: 3.25 } }}>
            <Stack spacing={{ xs: 2, md: 2.75 }}>
              <Stack
                direction={{ xs: "column", sm: "row" }}
                justifyContent="space-between"
                alignItems={{ xs: "flex-start", sm: "center" }}
                spacing={1}
              >
                <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
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
                  {roundLabel ? (
                    <Chip size="small" variant="outlined" label={roundLabel} />
                  ) : null}
                </Stack>

                {workoutName ? (
                  <Typography variant="body2" color="text.secondary">
                    {workoutName}
                  </Typography>
                ) : null}
              </Stack>

              <Stack
                spacing={{ xs: 1.5, md: 2.25 }}
                sx={{ minHeight: { md: 292 } }}
                justifyContent="space-between"
              >
                <Stack spacing={1.25}>
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
                        maxWidth: "12ch",
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
                        sx={{ mt: 1 }}
                      >
                        {heroMetric}
                      </Typography>
                    ) : null}
                  </Box>
                </Stack>

                <Box
                  sx={{
                    maxWidth: 620,
                    px: 1.5,
                    py: 1.25,
                    borderRadius: 3,
                    bgcolor: alpha(
                      currentTone === "pause"
                        ? theme.palette.warning.main
                        : theme.palette.common.black,
                      currentTone === "pause" ? 0.08 : 0.12,
                    ),
                  }}
                >
                  <Typography variant="body2" color="text.secondary">
                    {heroSupport || UI_TEXT.training.states.startToSeeSets}
                  </Typography>
                </Box>
              </Stack>

              {!training ? (
                <Typography color="text.secondary">
                  {UI_TEXT.training.states.startToSeeSets}
                </Typography>
              ) : null}
            </Stack>
          </CardContent>
        </Card>

        <Stack
          spacing={2}
          sx={{
            position: { lg: "sticky" },
            top: { lg: 104 },
          }}
        >
          <Card>
            <CardContent
              sx={{
                p: 2.25,
                minHeight: 228,
                display: "flex",
                flexDirection: "column",
              }}
            >
              <Typography variant="overline" color="text.secondary">
                {UI_TEXT.training.cards.next}
              </Typography>

              <Stack
                spacing={1.25}
                justifyContent="space-between"
                sx={{ flex: 1, pt: 1.25 }}
              >
                <Box>
                  <Typography variant="h5" sx={{ minHeight: 72 }}>
                    {nextPrimaryLabel}
                  </Typography>
                  <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ minHeight: 42, mt: 1 }}
                  >
                    {nextSupportLine}
                  </Typography>
                </Box>

                {nextNameStep?.estimatedSeconds ? (
                  <Chip
                    size="small"
                    variant="outlined"
                    sx={{ alignSelf: "flex-start" }}
                    label={formatCountdownMillis(
                      nextNameStep.estimatedSeconds * 1000,
                      { showHours },
                    )}
                  />
                ) : null}
              </Stack>
            </CardContent>
          </Card>

          {!isMobile ? (
            <Card>
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

      <Card>
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

            <Stack spacing={1}>
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
                  <Box
                    key={group.key}
                    sx={{
                      px: { xs: 1.5, md: 1.75 },
                      py: 1.5,
                      borderRadius: 3,
                      borderLeft: "3px solid",
                      borderLeftColor: group.current
                        ? "primary.main"
                        : "transparent",
                      bgcolor: group.current
                        ? alpha(theme.palette.primary.main, 0.08)
                        : alpha(theme.palette.common.black, 0.08),
                    }}
                  >
                    <Stack spacing={1}>
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
                          alignItems="center"
                        >
                          <Typography variant="subtitle1" fontWeight={700}>
                            {groupTitle}
                          </Typography>
                          {group.current ? (
                            <Chip size="small" color="primary" label="Current" />
                          ) : null}
                        </Stack>

                        <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
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
                          {group.hasSuperset ? (
                            <Chip size="small" variant="outlined" label="Superset" />
                          ) : null}
                        </Stack>
                      </Stack>

                      {group.subsets.length ? (
                        <Stack direction="row" spacing={0.75} useFlexGap flexWrap="wrap">
                          {group.subsets.flatMap((subset) =>
                            buildExercisePills(subset.exercises),
                          ).map((text, idx) => (
                            <Chip
                              key={`${group.key}-${idx}`}
                              size="small"
                              label={text}
                              sx={{
                                bgcolor: alpha(theme.palette.common.white, 0.04),
                              }}
                            />
                          ))}
                        </Stack>
                      ) : group.type === STEP_TYPE_PAUSE ? null : (
                        <Typography variant="body2" color="text.secondary">
                          {UI_TEXT.training.states.noExercises}
                        </Typography>
                      )}
                    </Stack>
                  </Box>
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
              bgcolor: alpha(theme.palette.background.paper, 0.92),
              backdropFilter: "blur(16px)",
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
