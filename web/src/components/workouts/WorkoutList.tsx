import type { AskConfirmOptions, Workout } from "../../types";
import { useWorkoutActions } from "../../hooks/useWorkoutActions";

export type WorkoutsListProps = {
  workouts: Workout[];
  loading?: boolean;

  // auth
  currentUserId: string | null;

  // state from parent
  selectedWorkoutId: string | null;
  setSelectedWorkoutId: (id: string | null) => void;

  // editor hooks (parent owns editor)
  editingWorkout: Workout | null;
  onNew: () => void;
  onEdit: (workoutId: string) => void;
  onOpenEditor: () => void;

  // data updates
  setWorkouts: (updater: (prev: Workout[] | null) => Workout[] | null) => void;

  // dialogs
  askConfirm: (
    message: string,
    options?: AskConfirmOptions,
  ) => Promise<boolean>;
  askPrompt: (message: string, defaultValue?: string) => Promise<string | null>;
  notifyUser: (message: string) => Promise<void>;

  // templates
  templatesReload: () => void;
};

export function WorkoutsList({
  workouts,
  loading,
  currentUserId,
  selectedWorkoutId,
  setSelectedWorkoutId,
  editingWorkout,
  onNew,
  onEdit,
  onOpenEditor,
  setWorkouts,
  askConfirm,
  askPrompt,
  notifyUser,
  templatesReload,
}: WorkoutsListProps) {
  // List actions (delete/share) still live here, but "new/edit" is delegated to parent
  const { removeWorkout, shareWorkout } = useWorkoutActions({
    workouts,
    editingWorkout,
    selectedWorkoutId,
    setEditingWorkout: () => {},
    setShowWorkoutForm: () => {},
    setSelectedWorkoutId,
    setWorkouts,
    askConfirm,
    askPrompt,
    notify: notifyUser,
    templatesReload,
  });

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
                  onClick={() => shareWorkout(workout.id)}
                >
                  Share
                </button>

                <button
                  className="btn subtle"
                  type="button"
                  onClick={() => removeWorkout(workout.id)}
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
