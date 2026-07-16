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

  onShare: (workoutId: string) => void;
  onDelete: (workoutId: string) => void;
};

export function WorkoutsList({
  workouts,
  loading,
  currentUserId,
  setSelectedWorkoutId,
  onNew,
  onEdit,
  onOpenEditor,
  onShare,
  onDelete,
}: WorkoutsListProps) {
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
          {workouts.map((workout) => (
            <li key={workout.id} className="list-item list-row">
              <div>
                <strong>{workout.name}</strong>
                <div className="muted small">{workout.steps.length} steps</div>
              </div>

              <div className="btn-group">
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
                  onClick={() => onShare(workout.id)}
                >
                  Share
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
