import type { Workout } from "../../types";

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
            title={!currentUserId ? "Login/select a user first" : ""}
          >
            New workout
          </button>
        </div>
      </div>

      {loading ? <p className="muted small">Loading workouts…</p> : null}

      {!loading && workouts.length === 0 ? (
        <p className="muted small">No workouts yet.</p>
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
