import type { CatalogExercise } from "../../types";
import { UI_TEXT } from "../../utils/uiText";

export type ExercisesViewData = {
  exercises: CatalogExercise[];
  isAdmin: boolean;
};

export type ExercisesViewActions = {
  onAddExercise: () => void | Promise<void>;
  onAddCoreExercise: () => void | Promise<void>;
  onRenameExercise: (exercise: CatalogExercise) => void | Promise<void>;
  onDeleteExercise: (exercise: CatalogExercise) => void | Promise<void>;
};

// ExercisesView manages personal and core exercises.
export function ExercisesView({
  data,
  actions,
}: {
  data: ExercisesViewData;
  actions: ExercisesViewActions;
}) {
  const { exercises, isAdmin } = data;
  const {
    onAddExercise,
    onAddCoreExercise,
    onRenameExercise,
    onDeleteExercise,
  } = actions;
  return (
    <section className="panel">
      <div className="panel-header">
        <div>
          <h3>{UI_TEXT.pages.exercises.title}</h3>
          <p className="muted small">{UI_TEXT.pages.exercises.hint}</p>
        </div>
        <div className="btn-group">
          <button className="btn primary" onClick={() => onAddExercise()}>
            {UI_TEXT.pages.exercises.addExercise}
          </button>
          {isAdmin && (
            <button className="btn subtle" onClick={() => onAddCoreExercise()}>
              {UI_TEXT.pages.exercises.addCoreExercise}
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
                  {ex.isCore
                    ? UI_TEXT.exercises.core
                    : UI_TEXT.exercises.personal}
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
        {!exercises.length && (
          <p className="muted">{UI_TEXT.pages.exercises.empty}</p>
        )}
      </ul>
    </section>
  );
}
