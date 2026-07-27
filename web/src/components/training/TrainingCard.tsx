import { useCallback, useMemo, type RefObject } from "react";
import PauseCircleRoundedIcon from "@mui/icons-material/PauseCircleRounded";
import PlayArrowRoundedIcon from "@mui/icons-material/PlayArrowRounded";
import SkipNextRoundedIcon from "@mui/icons-material/SkipNextRounded";
import TaskAltRoundedIcon from "@mui/icons-material/TaskAltRounded";
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  LinearProgress,
  Stack,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import { alpha } from "@mui/material/styles";

import { getCountdownDisplayMillis } from "../../utils/countdown";
import { formatCountdownMillis, formatElapsedMillis } from "../../utils/format";
import { PROMPTS } from "../../utils/messages";
import { STEP_TYPE_PAUSE } from "../../utils/step";
import type { Exercise, TrainingState, TrainingStepState } from "../../types";
import {
  formatExerciseMetric,
  formatRoundValue,
  formatStepValue,
} from "../../utils/trainingCard";
import { getExercises, getStepName } from "../../utils/training";

type TrainingMode = "superset" | "timed" | "recovery" | "manual";

function getTrainingMode(step: TrainingStepState | null): TrainingMode {
  if (step?.type === STEP_TYPE_PAUSE) return "recovery";
  if (step?.superset) return "superset";
  if ((step?.estimatedSeconds || 0) > 0) return "timed";
  return "manual";
}

function getUpcomingStep(
  training: TrainingState | null,
): TrainingStepState | null {
  if (!training) return null;
  for (
    let index = Math.max(0, training.currentIndex + 1);
    index < training.steps.length;
    index += 1
  ) {
    if (!training.steps[index].completed) return training.steps[index];
  }
  return null;
}

function getModeLabel(mode: TrainingMode) {
  if (mode === "superset") return "Superset";
  if (mode === "timed") return "Timed round";
  if (mode === "recovery") return "Recovery";
  return "Round";
}

function getExerciseName(exercise: Exercise | undefined, fallback: string) {
  return exercise?.name?.trim() || fallback;
}

function ExerciseMetric({ exercise }: { exercise: Exercise }) {
  const metric = formatExerciseMetric(exercise);
  if (!metric) return null;
  return (
    <Typography variant="h6" color="text.secondary" sx={{ fontWeight: 600 }}>
      {metric}
    </Typography>
  );
}

function SupersetExerciseList({ exercises }: { exercises: Exercise[] }) {
  return (
    <Stack
      divider={
        <Box
          sx={{
            borderTop: "1px solid",
            borderColor: "divider",
          }}
        />
      }
      sx={{
        border: "1px solid",
        borderColor: "primary.main",
        borderRadius: 3,
        overflow: "hidden",
        bgcolor: (theme) => alpha(theme.palette.primary.main, 0.045),
      }}
    >
      {exercises.map((exercise, index) => (
        <Box
          key={`${exercise.name}-${exercise.side || ""}-${index}`}
          sx={{
            display: "grid",
            gridTemplateColumns: { xs: "36px minmax(0, 1fr)", sm: "48px 1fr" },
            gap: { xs: 1.25, sm: 2 },
            alignItems: "center",
            px: { xs: 1.5, sm: 2.25 },
            py: { xs: 1.5, sm: 1.8 },
          }}
        >
          <Typography
            variant="h5"
            color="primary.main"
            sx={{ fontWeight: 800, textAlign: "center" }}
          >
            {index + 1}
          </Typography>
          <Box sx={{ minWidth: 0 }}>
            <Typography
              variant="h5"
              sx={{
                fontWeight: 800,
                lineHeight: 1.15,
                textTransform: "uppercase",
                overflowWrap: "anywhere",
              }}
            >
              {exercise.name}
            </Typography>
            <ExerciseMetric exercise={exercise} />
          </Box>
        </Box>
      ))}
    </Stack>
  );
}

function NextStepCard({
  step,
  emphasized,
}: {
  step: TrainingStepState | null;
  emphasized: boolean;
}) {
  const exercises = getExercises(step);
  const primary = exercises[0];
  const title = step
    ? step.superset
      ? getStepName(step)
      : getExerciseName(primary, getStepName(step))
    : "Training complete";
  const detail = step?.superset
    ? exercises.map((exercise) => exercise.name).join(" • ")
    : primary
      ? formatExerciseMetric(primary)
      : step?.estimatedSeconds
        ? formatCountdownMillis(step.estimatedSeconds * 1000)
        : "";

  return (
    <Card
      variant="outlined"
      sx={{
        borderColor: emphasized ? "warning.main" : "divider",
        bgcolor: emphasized
          ? (theme) => alpha(theme.palette.warning.main, 0.08)
          : undefined,
      }}
    >
      <CardContent sx={{ p: 2.25 }}>
        <Typography
          variant="overline"
          color={emphasized ? "warning.main" : "text.secondary"}
        >
          Up next
        </Typography>
        <Typography
          variant="h5"
          sx={{
            mt: 0.75,
            fontWeight: 800,
            lineHeight: 1.15,
            overflowWrap: "anywhere",
          }}
        >
          {title}
        </Typography>
        {detail ? (
          <Typography
            variant="body1"
            color="text.secondary"
            sx={{ mt: 1, lineHeight: 1.5 }}
          >
            {detail}
          </Typography>
        ) : null}
        {step && !step.superset && step.estimatedSeconds ? (
          <Chip
            size="small"
            variant="outlined"
            sx={{ mt: 1.5 }}
            label={formatCountdownMillis(step.estimatedSeconds * 1000)}
          />
        ) : null}
      </CardContent>
    </Card>
  );
}

function TrainingActions({
  training,
  running,
  done,
  startLabel,
  advanceLabel,
  finishAction,
  mobile,
  onStart,
  onPause,
  onAdvance,
  runButtonRef,
  nextButtonRef,
}: {
  training: TrainingState | null;
  running?: boolean;
  done?: boolean;
  startLabel: string;
  advanceLabel: string;
  finishAction: boolean;
  mobile: boolean;
  onStart: () => void;
  onPause: () => void;
  onAdvance: () => void;
  runButtonRef?: RefObject<HTMLButtonElement>;
  nextButtonRef?: RefObject<HTMLButtonElement>;
}) {
  return (
    <Stack
      direction={mobile ? "row" : "column"}
      spacing={1.25}
      sx={{ width: "100%", p: mobile ? 1.25 : 0 }}
    >
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
          minHeight: mobile ? 64 : 58,
          fontSize: mobile ? "0.95rem" : "1rem",
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
          finishAction ? <TaskAltRoundedIcon /> : <SkipNextRoundedIcon />
        }
        onClick={onAdvance}
        disabled={!training || done || !training.startedAt}
        sx={{
          minHeight: mobile ? 64 : 56,
          fontSize: mobile ? "0.95rem" : "0.98rem",
        }}
      >
        {advanceLabel}
      </Button>
    </Stack>
  );
}

export function TrainingCard({
  training,
  currentStep,
  elapsed,
  workoutName: _workoutName,
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
  const mode = getTrainingMode(currentStep);
  const isSuperset = mode === "superset";
  const isRecovery = mode === "recovery";
  const isTimed = mode === "timed" || isRecovery;
  const isCountdown =
    Boolean(currentStep?.autoAdvance) ||
    Boolean(currentStep?.pauseOptions?.autoAdvance);
  const targetMillis = (currentStep?.estimatedSeconds || 0) * 1000;
  const remainingMillis =
    isCountdown && targetMillis > 0
      ? getCountdownDisplayMillis(targetMillis, elapsed)
      : 0;
  const clockText =
    isCountdown && targetMillis > 0
      ? formatCountdownMillis(remainingMillis, { showHours })
      : formatElapsedMillis(elapsed, { showHours });
  const progress =
    isTimed && targetMillis > 0
      ? Math.min(100, Math.max(0, (elapsed / targetMillis) * 100))
      : 0;
  const isTransitioning =
    Boolean(running) &&
    isCountdown &&
    remainingMillis > 0 &&
    remainingMillis <= 5000;
  const exercises = getExercises(currentStep);
  const primaryExercise = exercises[0];
  const upcomingStep = useMemo(() => getUpcomingStep(training), [training]);

  const totalSteps = training?.steps.length || 0;
  const currentNumber = training ? training.currentIndex + 1 : 0;
  const stepValue = formatStepValue(currentNumber, totalSteps);
  const roundValue = formatRoundValue(currentStep);
  const isLastStep =
    Boolean(training?.steps.length) &&
    training!.currentIndex >= training!.steps.length - 1;
  const hasProgress = training?.steps.some(
    (step) => (step.elapsedMillis || 0) > 0 || step.completed,
  );
  const hasStarted = Boolean(training?.running) || Boolean(hasProgress);
  const startLabel = running ? "Pause" : hasStarted ? "Continue" : "Start";
  const advanceLabel = isLastStep
    ? "Finish training"
    : isSuperset
      ? "Finish superset"
      : isCountdown
        ? "Skip"
        : "Next";
  const finishAction = isLastStep || isSuperset;

  const handleAdvance = useCallback(() => {
    onStopAudio?.();
    if (isLastStep) {
      onFinish();
      return;
    }
    onNext();
  }, [isLastStep, onFinish, onNext, onStopAudio]);

  const actions = (
    <TrainingActions
      training={training}
      running={running}
      done={done}
      startLabel={startLabel}
      advanceLabel={advanceLabel}
      finishAction={finishAction}
      mobile={isMobile}
      onStart={onStart}
      onPause={onPause}
      onAdvance={handleAdvance}
      runButtonRef={runButtonRef}
      nextButtonRef={nextButtonRef}
    />
  );

  return (
    <Stack spacing={2.5}>
      <Box
        sx={{
          display: "grid",
          gap: 2.5,
          gridTemplateColumns: { xs: "1fr", lg: "minmax(0, 1fr) 320px" },
          alignItems: "start",
        }}
      >
        <Card
          sx={{
            overflow: "hidden",
            background: `linear-gradient(150deg, ${alpha(
              isRecovery
                ? theme.palette.warning.main
                : theme.palette.primary.main,
              0.12,
            )}, ${alpha(theme.palette.background.paper, 0.98)})`,
          }}
        >
          {isTransitioning ? (
            <Box
              sx={{
                py: 1,
                px: 2,
                bgcolor: "warning.main",
                color: "warning.contrastText",
                textAlign: "center",
              }}
            >
              <Typography
                variant="subtitle1"
                sx={{
                  fontWeight: 900,
                  letterSpacing: "0.14em",
                  textTransform: "uppercase",
                }}
              >
                Get ready
              </Typography>
            </Box>
          ) : null}

          <CardContent sx={{ p: { xs: 2, sm: 2.75, md: 3.5 } }}>
            <Stack spacing={{ xs: 2, md: 2.75 }}>
              <Stack
                direction="row"
                spacing={1}
                useFlexGap
                sx={{
                  flexWrap: "wrap",
                  alignItems: "center",
                  justifyContent: "space-between",
                }}
              >
                <Chip
                  label={getModeLabel(mode)}
                  color={isRecovery ? "warning" : "primary"}
                  sx={{
                    fontWeight: 800,
                    textTransform: "uppercase",
                    letterSpacing: "0.08em",
                  }}
                />
                <Stack direction="row" spacing={1}>
                  {roundValue ? (
                    <Chip variant="outlined" label={`Round ${roundValue}`} />
                  ) : null}
                  {stepValue ? (
                    <Chip variant="outlined" label={stepValue} />
                  ) : null}
                </Stack>
              </Stack>

              <Box sx={{ textAlign: "center", py: { xs: 0.5, md: 1 } }}>
                <Typography
                  variant={isMobile ? "h1" : "h1"}
                  sx={{
                    fontSize: {
                      xs: "clamp(4rem, 17vw, 6.2rem)",
                      md: "clamp(5.5rem, 9vw, 8rem)",
                    },
                    fontVariantNumeric: "tabular-nums",
                    letterSpacing: "-0.065em",
                    lineHeight: 0.95,
                    color: isTransitioning ? "warning.main" : "text.primary",
                  }}
                >
                  {clockText}
                </Typography>
                <Typography
                  variant="body1"
                  color="text.secondary"
                  sx={{ mt: 1.25 }}
                >
                  {isSuperset
                    ? "Elapsed · no time limit"
                    : isCountdown
                      ? `${formatCountdownMillis(targetMillis, {
                          showHours,
                        })} · advances automatically`
                      : targetMillis > 0
                        ? `Target ${formatCountdownMillis(targetMillis, {
                            showHours,
                          })}`
                        : "Elapsed"}
                </Typography>
              </Box>

              {isTimed && targetMillis > 0 ? (
                <LinearProgress
                  variant="determinate"
                  value={progress}
                  color={isRecovery ? "warning" : "primary"}
                  sx={{
                    height: 10,
                    borderRadius: 999,
                    bgcolor: alpha(
                      isRecovery
                        ? theme.palette.warning.main
                        : theme.palette.primary.main,
                      0.14,
                    ),
                  }}
                />
              ) : null}

              {!currentStep ? (
                <Typography
                  variant="h3"
                  sx={{ textAlign: "center", py: 5, fontWeight: 800 }}
                >
                  {PROMPTS.noTraining}
                </Typography>
              ) : isSuperset ? (
                <Stack spacing={1.5}>
                  <SupersetExerciseList exercises={exercises} />
                  <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ textAlign: "center" }}
                  >
                    Perform these exercises continuously. Finish the block when
                    you are done.
                  </Typography>
                </Stack>
              ) : isRecovery ? (
                <Box sx={{ textAlign: "center", py: { xs: 1.5, md: 3 } }}>
                  <Typography
                    variant="h2"
                    sx={{
                      fontWeight: 900,
                      textTransform: "uppercase",
                      letterSpacing: "0.04em",
                    }}
                  >
                    Rest
                  </Typography>
                  <Typography color="text.secondary" sx={{ mt: 1 }}>
                    {isCountdown
                      ? "The next round starts automatically."
                      : "Continue when you are ready."}
                  </Typography>
                </Box>
              ) : (
                <Box sx={{ textAlign: "center", py: { xs: 1.5, md: 3 } }}>
                  <Typography
                    variant="h2"
                    sx={{
                      fontSize: {
                        xs: "clamp(2.5rem, 10vw, 4rem)",
                        md: "clamp(3.5rem, 6vw, 5.5rem)",
                      },
                      fontWeight: 900,
                      lineHeight: 1,
                      textTransform: "uppercase",
                      overflowWrap: "anywhere",
                    }}
                  >
                    {getExerciseName(primaryExercise, getStepName(currentStep))}
                  </Typography>
                  {primaryExercise ? (
                    <Box sx={{ mt: 1.5 }}>
                      <ExerciseMetric exercise={primaryExercise} />
                    </Box>
                  ) : null}
                </Box>
              )}
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
          <NextStepCard step={upcomingStep} emphasized={isTransitioning} />
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
                {actions}
              </CardContent>
            </Card>
          ) : null}
        </Stack>
      </Box>

      {isMobile ? (
        <Box
          sx={{
            position: "sticky",
            bottom: 12,
            zIndex: 9,
          }}
        >
          <Card
            sx={{
              bgcolor: alpha(theme.palette.background.paper, 0.92),
              backdropFilter: "blur(16px)",
            }}
          >
            {actions}
          </Card>
        </Box>
      ) : null}
    </Stack>
  );
}
