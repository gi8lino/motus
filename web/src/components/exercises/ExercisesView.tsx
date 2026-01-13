import type { CatalogExercise } from "../../types";

// ExercisesView manages personal and core exercises.
export function ExercisesView({
  exercises,
  isAdmin,
  onAddExercise,
  onAddCoreExercise,
  onRenameExercise,
  onDeleteExercise,
}: {
  exercises: CatalogExercise[];
  isAdmin: boolean;
  onAddExercise: () => void | Promise<void>;
  onAddCoreExercise: () => void | Promise<void>;
  onRenameExercise: (exercise: CatalogExercise) => void | Promise<void>;
  onDeleteExercise: (exercise: CatalogExercise) => void | Promise<void>;
}) {
  return (
    <section className="panel">
      <div className="panel-header">
        <div>
          <h3>Exercises</h3>
          <p className="muted small">Manage reusable exercises.</p>
        </div>
        <div className="btn-group">
          <button className="btn primary" onClick={() => onAddExercise()}>
            Add Exercise
          </button>
          {isAdmin && (
            <button className="btn subtle" onClick={() => onAddCoreExercise()}>
              Add Core Exercise
            </button>
          )}
        </div>
      </div>
      {/* Exercise list */}
      <ul className="list">
        {exercises.map((ex) => (
          <li key={ex.id} className="list-item">
            <div className="list-row">
              <div>
                <strong>{ex.name}</strong>
                <span className={`exercise-tag ${ex.isCore ? "core" : "user"}`}>
                  {ex.isCore ? "Core" : "Personal"}
                </span>
                <div className="muted small">
                  {ex.createdAt
                    ? new Date(ex.createdAt).toLocaleDateString()
                    : ""}
                </div>
              </div>
              <div className="btn-group">
                {isAdmin && (
                  <>
                    <button
                      className="btn subtle"
                      onClick={() => onRenameExercise(ex)}
                    >
                      Rename
                    </button>
                    <button
                      className="btn subtle"
                      onClick={() => onDeleteExercise(ex)}
                    >
                      Delete
                    </button>
                  </>
                )}
              </div>
            </div>
          </li>
        ))}
        {!exercises.length && <p className="muted">No exercises yet.</p>}
      </ul>
    </section>
  );
}
