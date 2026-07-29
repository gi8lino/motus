import type { Workout } from "../../types";
import { UI_TEXT } from "../../utils/uiText";
import { EmptyState } from "../common/EmptyState";
import { LoadingState } from "../common/LoadingState";

export type WorkoutsListProps = {
  workouts: Workout[];
  loading?: boolean;

  // auth
  currentUserId: string | null;

  // state from parent
  setSelectedWorkoutId: (id: string | null) => void;

  // editor hooks (parent owns editor)
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
  return (
    <section className="panel">
      <div className="panel-header">
        <h3 style={{ margin: 0 }}>Workouts</h3>

        <div className="btn-group">
          <button
            className="btn primary"
            type="button"
            onClick={() => {
              if (!currentUserId) return;
              setSelectedWorkoutId(null);
              onNew();
              onOpenEditor();
            }}
            disabled={!currentUserId}
            title={!currentUserId ? UI_TEXT.prompts.selectUserFirst : ""}
          >
            New workout
          </button>
        </div>
      </div>

      {loading ? <LoadingState /> : null}

      <div className="list-row">
        <input
          aria-label="Search workouts"
          placeholder="Search workouts or tags"
          value={query}
          onChange={(event) => setQuery(event.target.value)}
        />
        <select
          aria-label="Sort workouts"
          value={sort}
          onChange={(event) =>
            setSort(event.target.value as "favorite" | "name" | "newest")
          }
        >
          <option value="favorite">Favorites first</option>
          <option value="newest">Newest</option>
          <option value="name">Name</option>
        </select>
      </div>

      {!loading && workouts.length === 0 ? (
        <EmptyState
          title="Build your first workout"
          description="Combine exercises, pauses, and repeats into a plan you can run anytime."
          actionLabel="Create workout"
          onAction={() => {
            if (!currentUserId) return;
            setSelectedWorkoutId(null);
            onNew();
            onOpenEditor();
          }}
        />
      ) : null}

      {workouts.length ? (
        <ul className="list">
          {visibleWorkouts.map((workout) => (
            <li key={workout.id} className="list-item list-row">
              <div>
                <strong>{workout.name}</strong>
                <div className="muted small">{workout.steps.length} steps</div>
                {workout.tags?.length ? (
                  <div className="muted small">{workout.tags.join(" • ")}</div>
                ) : null}
              </div>

              <div className="btn-group">
                <button
                  className="btn subtle"
                  type="button"
                  onClick={() => onFavorite(workout)}
                  aria-label={`${workout.favorite ? "Unfavorite" : "Favorite"} ${workout.name}`}
                >
                  {workout.favorite ? "★" : "☆"}
                </button>
                <button
                  className="btn subtle"
                  type="button"
                  onClick={() => onTags(workout)}
                >
                  Tags
                </button>
                <button
                  className="btn subtle"
                  type="button"
                  onClick={() => {
                    setSelectedWorkoutId(workout.id);
                    onEdit(workout.id);
                    onOpenEditor();
                  }}
                >
                  Edit
                </button>

                <button
                  className="btn subtle"
                  type="button"
                  onClick={() => onDuplicate(workout.id)}
                >
                  Duplicate
                </button>

                <button
                  className="btn subtle"
                  type="button"
                  onClick={() => onDelete(workout.id)}
                >
                  Delete
                </button>
              </div>
            </li>
          ))}
        </ul>
      ) : null}
    </section>
  );
}
import { useMemo, useState } from "react";
