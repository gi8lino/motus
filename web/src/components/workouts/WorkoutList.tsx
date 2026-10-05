import { useMemo, useState } from "react";
import ContentCopyRoundedIcon from "@mui/icons-material/ContentCopyRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditRoundedIcon from "@mui/icons-material/EditRounded";
import LocalOfferOutlinedIcon from "@mui/icons-material/LocalOfferOutlined";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import StarBorderRoundedIcon from "@mui/icons-material/StarBorderRounded";
import StarRoundedIcon from "@mui/icons-material/StarRounded";
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  IconButton,
  InputAdornment,
  MenuItem,
  Stack,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";

import type { Workout } from "../../types";
import { UI_TEXT } from "../../utils/uiText";
import { EmptyState } from "../common/EmptyState";
import { LoadingState } from "../common/LoadingState";

export type WorkoutsListProps = {
  workouts: Workout[];
  loading?: boolean;
  currentUserId: string | null;
  setSelectedWorkoutId: (id: string | null) => void;
  onNew: () => void;
  onEdit: (workoutId: string) => void;
  onOpenEditor: () => void;
  onDuplicate: (workoutId: string) => void;
  onDelete: (workoutId: string) => void;
  onFavorite: (workout: Workout) => void;
  onTags: (workout: Workout) => void;
};

export function WorkoutsList({
  workouts,
  loading,
  currentUserId,
  setSelectedWorkoutId,
  onNew,
  onEdit,
  onOpenEditor,
  onDuplicate,
  onDelete,
  onFavorite,
  onTags,
}: WorkoutsListProps) {
  const [query, setQuery] = useState("");
  const [sort, setSort] = useState<"favorite" | "name" | "newest">("favorite");
  const visibleWorkouts = useMemo(() => {
    const normalized = query.trim().toLowerCase();
    return workouts
      .filter(
        (workout) =>
          !normalized ||
          workout.name.toLowerCase().includes(normalized) ||
          workout.tags?.some((tag) => tag.toLowerCase().includes(normalized)),
      )
      .sort((a, b) => {
        if (sort === "name") return a.name.localeCompare(b.name);
        if (sort === "newest")
          return String(b.createdAt || "").localeCompare(
            String(a.createdAt || ""),
          );
        return Number(Boolean(b.favorite)) - Number(Boolean(a.favorite));
      });
  }, [query, sort, workouts]);
  const createWorkout = () => {
    if (!currentUserId) return;
    setSelectedWorkoutId(null);
    onNew();
    onOpenEditor();
  };

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
              <Typography variant="h4">Workouts</Typography>
              <Typography variant="body2" color="text.secondary">
                Build, organize, and reuse your training plans.
              </Typography>
            </Box>
            <Button
              variant="contained"
              size="large"
              onClick={createWorkout}
              disabled={!currentUserId}
              title={!currentUserId ? UI_TEXT.prompts.selectUserFirst : ""}
            >
              New workout
            </Button>
          </Stack>
          <Stack direction={{ xs: "column", sm: "row" }} spacing={1.25}>
            <TextField
              fullWidth
              size="small"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search workouts or tags"
              slotProps={{
                htmlInput: { "aria-label": "Search workouts" },
                input: {
                  startAdornment: (
                    <InputAdornment position="start">
                      <SearchRoundedIcon fontSize="small" />
                    </InputAdornment>
                  ),
                },
              }}
            />
            <TextField
              select
              size="small"
              value={sort}
              onChange={(event) => setSort(event.target.value as typeof sort)}
              label="Sort"
              sx={{ minWidth: { sm: 172 } }}
            >
              <MenuItem value="favorite">Favorites first</MenuItem>
              <MenuItem value="newest">Newest</MenuItem>
              <MenuItem value="name">Name</MenuItem>
            </TextField>
          </Stack>
          {loading ? <LoadingState rows={5} /> : null}
          {!loading && workouts.length === 0 ? (
            <EmptyState
              title="Build your first workout"
              description="Combine exercises, pauses, and repeats into a plan you can run anytime."
              actionLabel="Create workout"
              onAction={createWorkout}
            />
          ) : null}
          {!loading && visibleWorkouts.length === 0 && workouts.length > 0 ? (
            <EmptyState
              title="No matching workouts"
              description="Try another name or tag."
            />
          ) : null}
          <Stack spacing={1.25}>
            {visibleWorkouts.map((workout) => (
              <Card
                key={workout.id}
                variant="outlined"
                sx={{
                  boxShadow: "none",
                  transition: "transform 160ms ease, border-color 160ms ease",
                  "&:hover": {
                    transform: "translateY(-2px)",
                    borderColor: "primary.main",
                  },
                }}
              >
                <CardContent
                  sx={{
                    p: { xs: 1.75, sm: 2 },
                    "&:last-child": { pb: { xs: 1.75, sm: 2 } },
                  }}
                >
                  <Stack
                    direction={{ xs: "column", md: "row" }}
                    spacing={1.5}
                    sx={{
                      alignItems: { md: "center" },
                      justifyContent: "space-between",
                    }}
                  >
                    <Box sx={{ minWidth: 0 }}>
                      <Stack
                        direction="row"
                        spacing={0.75}
                        sx={{ alignItems: "center" }}
                      >
                        <Typography variant="h6" noWrap>
                          {workout.name}
                        </Typography>
                        {workout.favorite ? (
                          <StarRoundedIcon
                            color="warning"
                            fontSize="small"
                            aria-label="Favorite"
                          />
                        ) : null}
                      </Stack>
                      <Typography variant="body2" color="text.secondary">
                        {workout.steps.length}{" "}
                        {workout.steps.length === 1 ? "step" : "steps"}
                      </Typography>
                      {workout.tags?.length ? (
                        <Stack
                          direction="row"
                          spacing={0.5}
                          useFlexGap
                          sx={{ mt: 1, flexWrap: "wrap" }}
                        >
                          {workout.tags.map((tag) => (
                            <Chip key={tag} size="small" label={tag} />
                          ))}
                        </Stack>
                      ) : null}
                    </Box>
                    <Stack
                      direction="row"
                      spacing={0.25}
                      sx={{ alignSelf: { xs: "flex-end", md: "center" } }}
                    >
                      <Tooltip
                        title={
                          workout.favorite ? "Remove favorite" : "Favorite"
                        }
                      >
                        <IconButton
                          color={workout.favorite ? "warning" : "default"}
                          onClick={() => onFavorite(workout)}
                          aria-label={`${workout.favorite ? "Unfavorite" : "Favorite"} ${workout.name}`}
                        >
                          {workout.favorite ? (
                            <StarRoundedIcon />
                          ) : (
                            <StarBorderRoundedIcon />
                          )}
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Edit tags">
                        <IconButton
                          onClick={() => onTags(workout)}
                          aria-label={`Edit tags for ${workout.name}`}
                        >
                          <LocalOfferOutlinedIcon />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Edit workout">
                        <IconButton
                          color="primary"
                          onClick={() => {
                            setSelectedWorkoutId(workout.id);
                            onEdit(workout.id);
                            onOpenEditor();
                          }}
                          aria-label={`Edit ${workout.name}`}
                        >
                          <EditRoundedIcon />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Duplicate workout">
                        <IconButton
                          onClick={() => onDuplicate(workout.id)}
                          aria-label={`Duplicate ${workout.name}`}
                        >
                          <ContentCopyRoundedIcon />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Delete workout">
                        <IconButton
                          color="error"
                          onClick={() => onDelete(workout.id)}
                          aria-label={`Delete ${workout.name}`}
                        >
                          <DeleteOutlineRoundedIcon />
                        </IconButton>
                      </Tooltip>
                    </Stack>
                  </Stack>
                </CardContent>
              </Card>
            ))}
          </Stack>
        </Stack>
      </CardContent>
    </Card>
  );
}
