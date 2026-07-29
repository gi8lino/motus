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
  TextField,
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
  formatExerciseSide,
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
  if (mode === "timed") return "Timed exercise";
  if (mode === "recovery") return "Recovery";
  return "Exercise";
}

function getExerciseName(exercise: Exercise | undefined, fallback: string) {
  return exercise?.name?.trim() || fallback;
}

function getExerciseNameWithSide(exercise: Exercise) {
  const side = formatExerciseSide(exercise);
  return side ? `${exercise.name} (${side})` : exercise.name;
}

function SideBadge({ exercise }: { exercise: Exercise | undefined }) {
  const side = formatExerciseSide(exercise);
  if (!side) return null;
  return (
    <Chip
      size="small"
      color="primary"
      variant="outlined"
      label={side}
      sx={{
        height: 28,
        fontWeight: 900,
        letterSpacing: "0.09em",
        textTransform: "uppercase",
      }}
    />
  );
}

function ExerciseMetric({ exercise }: { exercise: Exercise }) {
  const metric = formatExerciseMetric(exercise);
  if (!metric) return null;
  return (
    <Typography variant="h6" color="text.secondary" sx={{ fontWeight: 650 }}>
      {metric}
    </Typography>
  );
}

function SupersetExerciseList({ exercises }: { exercises: Exercise[] }) {
  return (
    <Stack
      divider={<Box sx={{ borderTop: "1px solid", borderColor: "divider" }} />}
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
            gridTemplateColumns: { xs: "34px minmax(0, 1fr)", sm: "46px 1fr" },
            gap: { xs: 1.25, sm: 2 },
            alignItems: "center",
            px: { xs: 1.4, sm: 2.25 },
            py: { xs: 1.45, sm: 1.75 },
          }}
        >
          <Typography
            variant="h5"
            color="primary.main"
            sx={{ fontWeight: 900, textAlign: "center" }}
          >
            {index + 1}
          </Typography>

          <Stack spacing={0.6} sx={{ minWidth: 0 }}>
            <Stack
              direction="row"
              spacing={1}
              useFlexGap
              sx={{ alignItems: "center", flexWrap: "wrap" }}
            >
              <Typography
                variant="h5"
                sx={{
                  fontWeight: 850,
                  lineHeight: 1.15,
                  textTransform: "uppercase",
                  overflowWrap: "anywhere",
                }}
              >
                {exercise.name}
              </Typography>
              <SideBadge exercise={exercise} />
            </Stack>
            <ExerciseMetric exercise={exercise} />
          </Stack>
        </Box>
      ))}
    </Stack>
  );
}

function TimerDisplay({
  clockText,
  supportText,
  transitioning,
  compact,
}: {
  clockText: string;
  supportText: string;
  transitioning: boolean;
  compact?: boolean;
}) {
  return (
    <Box sx={{ textAlign: "center", py: compact ? 0.5 : { xs: 1, md: 1.5 } }}>
      <Typography
        variant="h1"
        sx={{
          fontSize: compact
            ? { xs: "clamp(3.6rem, 15vw, 5.4rem)", md: "5.5rem" }
            : {
                xs: "clamp(4rem, 17vw, 6.2rem)",
                md: "clamp(5.5rem, 9vw, 8rem)",
              },
          fontVariantNumeric: "tabular-nums",
          letterSpacing: "-0.065em",
          lineHeight: 0.95,
          color: transitioning ? "warning.main" : "text.primary",
        }}
      >
        {clockText}
      </Typography>
      <Typography variant="body1" color="text.secondary" sx={{ mt: 1.1 }}>
        {supportText}
      </Typography>
    </Box>
  );
}

function Timeline({ value, recovery }: { value: number; recovery: boolean }) {
  return (
    <LinearProgress
      aria-label="Timer progress"
      variant="determinate"
      value={value}
      color={recovery ? "warning" : "primary"}
      sx={{
        height: 10,
        borderRadius: 999,
        bgcolor: (theme) =>
          alpha(
            recovery ? theme.palette.warning.main : theme.palette.primary.main,
            0.14,
          ),
      }}
    />
  );
}

function NextStepPreview({
  step,
  emphasized,
}: {
  step: TrainingStepState | null;
  emphasized: boolean;
}) {
  const exercises = getExercises(step);
  const primary = exercises[0];
  const title = step
    ? step.superset && exercises.length > 1
      ? getStepName(step)
      : getExerciseName(primary, getStepName(step))
    : "Training complete";
  const side = formatExerciseSide(primary);
  const detail =
    step?.superset && exercises.length > 1
      ? exercises.map(getExerciseNameWithSide).join(" • ")
      : primary
        ? [formatExerciseMetric(primary), side].filter(Boolean).join(" · ")
        : step?.estimatedSeconds
          ? formatCountdownMillis(step.estimatedSeconds * 1000)
          : "";

  return (
    <Box
      sx={{
        pt: 2.25,
        borderTop: "1px solid",
        borderColor: emphasized ? "warning.main" : "divider",
      }}
    >
      <Stack
        direction={{ xs: "column", sm: "row" }}
        spacing={1}
        sx={{ justifyContent: "space-between", alignItems: "flex-start" }}
      >
        <Box sx={{ minWidth: 0 }}>
          <Typography
            variant="overline"
            color={emphasized ? "warning.main" : "text.secondary"}
          >
            Up next
          </Typography>
          <Typography
            variant="h6"
            sx={{
              mt: 0.25,
              fontWeight: 850,
              lineHeight: 1.2,
              overflowWrap: "anywhere",
            }}
          >
            {title}
          </Typography>
          {detail ? (
            <Typography color="text.secondary" sx={{ mt: 0.45 }}>
              {detail}
            </Typography>
          ) : null}
        </Box>

        {step?.estimatedSeconds ? (
          <Chip
            size="small"
            variant="outlined"
            label={formatCountdownMillis(step.estimatedSeconds * 1000)}
          />
        ) : null}
      </Stack>
    </Box>
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
      direction="row"
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
          minHeight: mobile ? 64 : 58,
          fontSize: mobile ? "0.95rem" : "1rem",
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
  onUpdateExercisePerformance,
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
  onUpdateExercisePerformance?: (
    exerciseIndex: number,
    patch: { actualReps?: string; actualWeight?: string },
  ) => void;
}) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down("md"));
  const running = training?.running;
  const done = training?.done;
  const mode = getTrainingMode(currentStep);
  const isSuperset = mode === "superset";
  const isRecovery = mode === "recovery";
  const exercises = getExercises(currentStep);
  const primaryExercise = exercises[0];
  const isCountdown =
    Boolean(currentStep?.autoAdvance) ||
    Boolean(currentStep?.pauseOptions?.autoAdvance);
  const targetMillis = (currentStep?.estimatedSeconds || 0) * 1000;
  const hasTimerTarget = targetMillis > 0;
  const remainingMillis =
    isCountdown && hasTimerTarget
      ? getCountdownDisplayMillis(targetMillis, elapsed)
      : 0;
  const clockText =
    isCountdown && hasTimerTarget
      ? formatCountdownMillis(remainingMillis, { showHours })
      : formatElapsedMillis(elapsed, { showHours });
  const progress = hasTimerTarget
    ? Math.min(100, Math.max(0, (elapsed / targetMillis) * 100))
    : 0;
  const isTransitioning =
    Boolean(running) &&
    isCountdown &&
    remainingMillis > 0 &&
    remainingMillis <= 5000;
  const timerSupport = isCountdown
    ? `${formatCountdownMillis(targetMillis, {
        showHours,
      })} · advances automatically`
    : hasTimerTarget
      ? `Target ${formatCountdownMillis(targetMillis, {
          showHours,
        })} · continues after cue`
      : "Elapsed";
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
    : isCountdown
      ? "Skip"
      : isSuperset
        ? "Finish superset"
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
    <Stack spacing={2}>
      <Card
        sx={{
          overflow: "hidden",
          background: `linear-gradient(150deg, ${alpha(
            isRecovery
              ? theme.palette.warning.main
              : theme.palette.primary.main,
            0.11,
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
          <Stack spacing={{ xs: 2.25, md: 2.75 }}>
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
                  fontWeight: 850,
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

            {!currentStep ? (
              <Typography
                variant="h3"
                sx={{ textAlign: "center", py: 6, fontWeight: 850 }}
              >
                {PROMPTS.noTraining}
              </Typography>
            ) : isSuperset ? (
              <Stack spacing={1.75}>
                <TimerDisplay
                  clockText={clockText}
                  supportText={
                    hasTimerTarget ? timerSupport : "Elapsed · no time limit"
                  }
                  transitioning={isTransitioning}
                  compact
                />
                <SupersetExerciseList exercises={exercises} />
                {hasTimerTarget ? (
                  <Timeline value={progress} recovery={false} />
                ) : null}
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
              <Stack spacing={1.75}>
                <Typography
                  variant="h2"
                  sx={{
                    textAlign: "center",
                    fontWeight: 900,
                    textTransform: "uppercase",
                    letterSpacing: "0.04em",
                  }}
                >
                  Rest
                </Typography>
                <TimerDisplay
                  clockText={clockText}
                  supportText={
                    isCountdown
                      ? "The next exercise starts automatically"
                      : "Continue when you are ready"
                  }
                  transitioning={isTransitioning}
                />
                {hasTimerTarget ? <Timeline value={progress} recovery /> : null}
              </Stack>
            ) : (
              <Stack spacing={1.8}>
                <Stack
                  spacing={1}
                  sx={{ alignItems: "center", textAlign: "center" }}
                >
                  <Stack
                    direction="row"
                    spacing={1.25}
                    useFlexGap
                    sx={{
                      alignItems: "center",
                      justifyContent: "center",
                      flexWrap: "wrap",
                    }}
                  >
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
                      {getExerciseName(
                        primaryExercise,
                        getStepName(currentStep),
                      )}
                    </Typography>
                    <SideBadge exercise={primaryExercise} />
                  </Stack>
                  {primaryExercise ? (
                    <ExerciseMetric exercise={primaryExercise} />
                  ) : null}
                </Stack>

                <TimerDisplay
                  clockText={clockText}
                  supportText={timerSupport}
                  transitioning={isTransitioning}
                />
                {hasTimerTarget ? (
                  <Timeline value={progress} recovery={false} />
                ) : null}
              </Stack>
            )}

            {currentStep && !isRecovery && exercises.length > 0 ? (
              <Stack direction="row" spacing={1}>
                <TextField
                  size="small"
                  label="Actual reps"
                  value={
                    primaryExercise?.actualReps ?? primaryExercise?.reps ?? ""
                  }
                  onChange={(event) =>
                    onUpdateExercisePerformance?.(0, {
                      actualReps: event.target.value,
                    })
                  }
                />
                <TextField
                  size="small"
                  label="Actual weight"
                  value={
                    primaryExercise?.actualWeight ??
                    primaryExercise?.weight ??
                    ""
                  }
                  onChange={(event) =>
                    onUpdateExercisePerformance?.(0, {
                      actualWeight: event.target.value,
                    })
                  }
                />
              </Stack>
            ) : null}

            <NextStepPreview step={upcomingStep} emphasized={isTransitioning} />

            {!isMobile ? (
              <Box
                sx={{
                  pt: 2.25,
                  borderTop: "1px solid",
                  borderColor: "divider",
                }}
              >
                {actions}
              </Box>
            ) : null}
          </Stack>
        </CardContent>
      </Card>

      {isMobile ? (
        <Box sx={{ position: "sticky", bottom: 12, zIndex: 9 }}>
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
