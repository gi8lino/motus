import FormatListNumberedRoundedIcon from "@mui/icons-material/FormatListNumberedRounded";
import PauseCircleOutlineRoundedIcon from "@mui/icons-material/PauseCircleOutlineRounded";
import RepeatRoundedIcon from "@mui/icons-material/RepeatRounded";
import { Box, Card, CardContent, Chip, Stack, Typography } from "@mui/material";

import type {
  Exercise,
  Workout,
  WorkoutStep,
  WorkoutSubset,
} from "../../types";
import { formatCountdownMillis } from "../../utils/format";
import { STEP_TYPE_PAUSE } from "../../utils/step";

function exerciseDetail(exercise: Exercise) {
  const detail = [exercise.reps, exercise.weight, exercise.duration]
    .map((value) => value?.trim())
    .filter(Boolean)
    .join(" · ");
  const side =
    exercise.side && exercise.side !== "not_applicable" ? exercise.side : "";
  return [detail, side].filter(Boolean).join(" · ");
}

function stepExercises(step: WorkoutStep) {
  if (step.subsets?.length) {
    return step.subsets.flatMap((subset) => subset.exercises || []);
  }
  return step.exercises || [];
}

function stepDuration(step: WorkoutStep, subsets: WorkoutSubset[]) {
  const seconds =
    step.estimatedSeconds ||
    subsets.reduce(
      (total, subset) => total + (subset.estimatedSeconds || 0),
      0,
    );
  return seconds > 0 ? formatCountdownMillis(seconds * 1000) : "At your pace";
}

export function WorkoutPlanPreview({ workout }: { workout: Workout }) {
  return (
    <Card
      variant="outlined"
      sx={{
        borderColor: "primary.main",
        bgcolor: (theme) => `${theme.palette.primary.main}08`,
        boxShadow: "none",
      }}
    >
      <CardContent
        sx={{
          p: { xs: 1.75, sm: 2.25 },
          "&:last-child": { pb: { xs: 1.75, sm: 2.25 } },
        }}
      >
        <Stack spacing={1.75}>
          <Stack
            direction={{ xs: "column", sm: "row" }}
            spacing={1}
            sx={{
              justifyContent: "space-between",
              alignItems: { sm: "center" },
            }}
          >
            <Box>
              <Typography variant="overline" color="primary.main">
                Selected workout
              </Typography>
              <Typography variant="h5">{workout.name}</Typography>
              <Typography variant="body2" color="text.secondary">
                Review the plan, then start when you are ready.
              </Typography>
            </Box>
            <Chip
              icon={<FormatListNumberedRoundedIcon />}
              color="primary"
              variant="outlined"
              label={`${workout.steps.length} ${workout.steps.length === 1 ? "step" : "steps"}`}
            />
          </Stack>

          <Stack spacing={1}>
            {workout.steps.map((step, index) => {
              const subsets = step.subsets || [];
              const exercises = stepExercises(step);
              const isPause = step.type === STEP_TYPE_PAUSE;
              return (
                <Box
                  key={step.id || `${step.name}-${index}`}
                  sx={{
                    display: "grid",
                    gridTemplateColumns: "32px minmax(0, 1fr)",
                    gap: 1.25,
                    alignItems: "start",
                    p: 1.25,
                    borderRadius: 2.5,
                    bgcolor: "background.paper",
                    border: 1,
                    borderColor: "divider",
                  }}
                >
                  <Typography
                    color="primary.main"
                    sx={{
                      fontWeight: 900,
                      textAlign: "center",
                      lineHeight: 1.75,
                    }}
                  >
                    {index + 1}
                  </Typography>
                  <Stack spacing={0.65} sx={{ minWidth: 0 }}>
                    <Stack
                      direction="row"
                      spacing={0.75}
                      useFlexGap
                      sx={{ alignItems: "center", flexWrap: "wrap" }}
                    >
                      {isPause ? (
                        <PauseCircleOutlineRoundedIcon
                          color="warning"
                          fontSize="small"
                        />
                      ) : null}
                      <Typography variant="subtitle1" sx={{ fontWeight: 750 }}>
                        {step.name || (isPause ? "Pause" : "Exercise")}
                      </Typography>
                      <Chip
                        size="small"
                        variant="outlined"
                        label={stepDuration(step, subsets)}
                      />
                      {(step.repeatCount || 1) > 1 ? (
                        <Chip
                          size="small"
                          icon={<RepeatRoundedIcon />}
                          label={`${step.repeatCount} rounds`}
                        />
                      ) : null}
                    </Stack>
                    {exercises.length ? (
                      <Stack spacing={0.3}>
                        {exercises.map((exercise, exerciseIndex) => (
                          <Typography
                            key={`${exercise.name}-${exerciseIndex}`}
                            variant="body2"
                            color="text.secondary"
                          >
                            {exercise.name}
                            {exerciseDetail(exercise)
                              ? ` · ${exerciseDetail(exercise)}`
                              : ""}
                          </Typography>
                        ))}
                      </Stack>
                    ) : isPause ? (
                      <Typography variant="body2" color="text.secondary">
                        {step.pauseOptions?.autoAdvance
                          ? "Automatically advances when complete."
                          : "Continue when you are ready."}
                      </Typography>
                    ) : null}
                  </Stack>
                </Box>
              );
            })}
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  );
}
