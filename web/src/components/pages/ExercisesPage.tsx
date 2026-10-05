import { useEffect, useMemo, useState } from "react";
import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import SwapHorizRoundedIcon from "@mui/icons-material/SwapHorizRounded";
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  IconButton,
  InputAdornment,
  Stack,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";

import type { CatalogExercise } from "../../types";
import { UI_TEXT } from "../../utils/uiText";
import { EmptyState } from "../common/EmptyState";
import { LoadingState } from "../common/LoadingState";

export type ExercisesViewData = {
  exercises: CatalogExercise[];
  isAdmin: boolean;
  loading: boolean;
};
export type ExercisesViewActions = {
  onAddExercise: () => void | Promise<void>;
  onAddCoreExercise: () => void | Promise<void>;
  onRenameExercise: (exercise: CatalogExercise) => void | Promise<void>;
  onDeleteExercise: (exercise: CatalogExercise) => void | Promise<void>;
  onToggleSides: (exercise: CatalogExercise) => void | Promise<void>;
};

export function ExercisesView({
  data,
  actions,
}: {
  data: ExercisesViewData;
  actions: ExercisesViewActions;
}) {
  const { exercises, isAdmin, loading } = data;
  const pageSize = 24;
  const [query, setQuery] = useState("");
  const [visibleCount, setVisibleCount] = useState(pageSize);
  const filteredExercises = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    return exercises.filter(
      (exercise) =>
        !normalized ||
        exercise.name.toLowerCase().includes(normalized) ||
        exercise.labels?.some((label) =>
          label.toLowerCase().includes(normalized),
        ),
    );
  }, [exercises, query]);
  const visibleExercises = filteredExercises.slice(0, visibleCount);
  useEffect(() => setVisibleCount(pageSize), [exercises.length, query]);

  return (
    <Card component="section">
      <CardContent
        sx={{ p: { xs: 2, sm: 3 }, "&:last-child": { pb: { xs: 2, sm: 3 } } }}
      >
        <Stack spacing={2.5}>
          <Stack
            direction={{ xs: "column", sm: "row" }}
            spacing={1.5}
            sx={{
              alignItems: { sm: "center" },
              justifyContent: "space-between",
            }}
          >
            <Box>
              <Typography variant="h4">
                {UI_TEXT.pages.exercises.title}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                {UI_TEXT.pages.exercises.hint}
              </Typography>
            </Box>
            <Stack direction={{ xs: "column", sm: "row" }} spacing={1}>
              <Button
                variant="contained"
                size="large"
                startIcon={<AddRoundedIcon />}
                onClick={() => void actions.onAddExercise()}
              >
                {UI_TEXT.pages.exercises.addExercise}
              </Button>
              {isAdmin ? (
                <Button
                  variant="outlined"
                  onClick={() => void actions.onAddCoreExercise()}
                >
                  Add core exercise
                </Button>
              ) : null}
            </Stack>
          </Stack>
          <TextField
            fullWidth
            size="small"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search exercises or labels"
            slotProps={{
              htmlInput: { "aria-label": "Search exercises" },
              input: {
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchRoundedIcon fontSize="small" />
                  </InputAdornment>
                ),
              },
            }}
          />
          {loading ? <LoadingState rows={6} /> : null}
          {!loading && exercises.length === 0 ? (
            <EmptyState
              title={UI_TEXT.pages.exercises.empty}
              description="Add exercises once, then reuse them in every workout."
              actionLabel={UI_TEXT.pages.exercises.addExercise}
              onAction={() => void actions.onAddExercise()}
            />
          ) : null}
          {!loading &&
          exercises.length > 0 &&
          filteredExercises.length === 0 ? (
            <EmptyState
              title="No matching exercises"
              description="Try another exercise name or label."
            />
          ) : null}
          <Stack spacing={1.1}>
            {visibleExercises.map((exercise) => (
              <Card
                key={exercise.id}
                variant="outlined"
                sx={{ boxShadow: "none" }}
              >
                <CardContent
                  sx={{
                    p: { xs: 1.5, sm: 1.75 },
                    "&:last-child": { pb: { xs: 1.5, sm: 1.75 } },
                  }}
                >
                  <Stack
                    direction={{ xs: "column", sm: "row" }}
                    spacing={1.25}
                    sx={{
                      alignItems: { sm: "center" },
                      justifyContent: "space-between",
                    }}
                  >
                    <Box>
                      <Stack
                        direction="row"
                        spacing={0.75}
                        useFlexGap
                        sx={{ alignItems: "center", flexWrap: "wrap" }}
                      >
                        <Typography
                          variant="subtitle1"
                          sx={{ fontWeight: 750 }}
                        >
                          {exercise.name}
                        </Typography>
                        <Chip
                          size="small"
                          color={exercise.isCore ? "warning" : "default"}
                          label={
                            exercise.isCore
                              ? UI_TEXT.exercises.core
                              : UI_TEXT.exercises.personal
                          }
                        />
                        {exercise.hasSides ? (
                          <Chip size="small" variant="outlined" label="Sides" />
                        ) : null}
                      </Stack>
                      {exercise.labels?.length ? (
                        <Stack
                          direction="row"
                          spacing={0.5}
                          useFlexGap
                          sx={{ mt: 0.75, flexWrap: "wrap" }}
                        >
                          {exercise.labels.map((label) => (
                            <Chip
                              key={label}
                              size="small"
                              variant="outlined"
                              label={label}
                            />
                          ))}
                        </Stack>
                      ) : null}
                    </Box>
                    {isAdmin || !exercise.isCore ? (
                      <Stack
                        direction="row"
                        spacing={0.25}
                        sx={{ alignSelf: { xs: "flex-end", sm: "center" } }}
                      >
                        <Tooltip
                          title={
                            exercise.hasSides ? "Disable sides" : "Enable sides"
                          }
                        >
                          <IconButton
                            onClick={() => void actions.onToggleSides(exercise)}
                            aria-label={`${exercise.hasSides ? "Disable" : "Enable"} sides for ${exercise.name}`}
                          >
                            <SwapHorizRoundedIcon />
                          </IconButton>
                        </Tooltip>
                        {isAdmin ? (
                          <>
                            <Tooltip title="Rename exercise">
                              <IconButton
                                color="primary"
                                onClick={() =>
                                  void actions.onRenameExercise(exercise)
                                }
                                aria-label={`Rename ${exercise.name}`}
                              >
                                <EditRoundedIcon />
                              </IconButton>
                            </Tooltip>
                            <Tooltip title="Delete exercise">
                              <IconButton
                                color="error"
                                onClick={() =>
                                  void actions.onDeleteExercise(exercise)
                                }
                                aria-label={`Delete ${exercise.name}`}
                              >
                                <DeleteOutlineRoundedIcon />
                              </IconButton>
                            </Tooltip>
                          </>
                        ) : null}
                      </Stack>
                    ) : null}
                  </Stack>
                </CardContent>
              </Card>
            ))}
          </Stack>
          {visibleCount < filteredExercises.length ? (
            <Box sx={{ textAlign: "center" }}>
              <Button
                variant="outlined"
                onClick={() =>
                  setVisibleCount((current) =>
                    Math.min(current + pageSize, filteredExercises.length),
                  )
                }
              >
                Show more ({filteredExercises.length - visibleCount} remaining)
              </Button>
            </Box>
          ) : null}
        </Stack>
      </CardContent>
    </Card>
  );
}
