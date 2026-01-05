import type { Workout } from "../types";

// WorkoutsView lists workouts and primary actions.
export function WorkoutsView({
  workouts,
  loading,
  authHeaderEnabled,
  currentUserName,
  hasUser,
  onNewWorkout,
  onEditWorkout,
  onShareTemplate,
  onDeleteWorkout,
}: {
  workouts: Workout[];
  loading: boolean;
  authHeaderEnabled: boolean;
  currentUserName: string;
  hasUser: boolean;
  onNewWorkout: () => void;
  onEditWorkout: (workoutId: string) => void;
  onShareTemplate: (workoutId: string) => void;
  onDeleteWorkout: (workoutId: string) => void;
}) {
  return (
    <section className="panel">
      <div className="panel-header">
        <h3>Workouts</h3>
        {authHeaderEnabled && (
          <div className="muted small">{currentUserName}</div>
        )}
        <button
          className="btn primary"
          onClick={onNewWorkout}
          disabled={!hasUser}
        >
          New Workout
        </button>
      </div>
      {loading && <p>Loading workouts…</p>}
      {/* Workout list */}
      <ul className="list">
        {workouts.map((w) => (
          <li key={w.id} className="list-item">
            <div>
              <div className="list-row">
                <div>
                  <strong>{w.name}</strong>
                  <div className="muted">
                    {w.steps.length} steps •{" "}
                    {w.createdAt
                      ? new Date(w.createdAt).toLocaleDateString()
                      : "recent"}
                  </div>
                  <div className="muted small">{stepsPreview(w)}</div>
                </div>
                <div className="btn-group">
                  <button
                    className="btn subtle"
                    onClick={() => onEditWorkout(w.id)}
                  >
                    Edit
                  </button>
                  <button
                    className="btn subtle"
                    onClick={() => onShareTemplate(w.id)}
                  >
                    Share
                  </button>
                  <button
                    className="btn subtle"
                    onClick={() => onDeleteWorkout(w.id)}
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          </li>
        ))}
      </ul>
    </section>
  );
}

// stepsPreview builds a compact preview of the first few steps.
function stepsPreview(workout: Workout) {
  if (!workout.steps?.length) return "No steps";
  const previews = workout.steps
    .slice(0, 3)
    .map((step) => step.name || step.type);
  return previews.join(" • ") + (workout.steps.length > 3 ? " …" : "");
}
