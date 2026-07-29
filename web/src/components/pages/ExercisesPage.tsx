import { useEffect, useMemo, useState } from "react";
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

// ExercisesView manages personal and core exercises.
export function ExercisesView({
  data,
  actions,
}: {
  data: ExercisesViewData;
  actions: ExercisesViewActions;
}) {
  const { exercises, isAdmin, loading } = data;
  const pageSize = 24;
  const [visibleCount, setVisibleCount] = useState(pageSize);
  const visibleExercises = useMemo(
    () => exercises.slice(0, visibleCount),
    [exercises, visibleCount],
  );

  useEffect(() => {
    setVisibleCount(pageSize);
  }, [exercises.length]);
  const {
    onAddExercise,
    onAddCoreExercise,
    onRenameExercise,
    onDeleteExercise,
    onToggleSides,
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
      {loading ? <LoadingState rows={6} /> : null}
      {/* Exercise list */}
      <ul className="list">
        {visibleExercises.map((ex) => (
          <li key={ex.id} className="list-item">
            <div className="list-row">
              <div>
                <strong>{ex.name}</strong>
                <span className={`exercise-tag ${ex.isCore ? "core" : "user"}`}>
                  {ex.isCore
                    ? UI_TEXT.exercises.core
                    : UI_TEXT.exercises.personal}
                </span>
                {ex.labels?.map((label) => (
                  <span key={label} className="exercise-label">
                    {label}
                  </span>
                ))}
                <div className="muted small">
                  {ex.createdAt
                    ? new Date(ex.createdAt).toLocaleDateString()
                    : ""}
                </div>
              </div>
              <div className="btn-group">
                {(!ex.isCore || isAdmin) && (
                  <button
                    className="btn subtle"
                    onClick={() => onToggleSides(ex)}
                  >
                    {ex.hasSides ? "Disable sides" : "Enable sides"}
                  </button>
                )}
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
        {!loading && !exercises.length && (
          <EmptyState
            title={UI_TEXT.pages.exercises.empty}
            description="Add exercises once, then reuse them in every workout."
            actionLabel={UI_TEXT.pages.exercises.addExercise}
            onAction={onAddExercise}
          />
        )}
      </ul>
      {visibleCount < exercises.length ? (
        <div className="actions">
          <button
            className="btn subtle"
            type="button"
            onClick={() =>
              setVisibleCount((current) =>
                Math.min(current + pageSize, exercises.length),
              )
            }
          >
            Show more ({exercises.length - visibleCount} remaining)
          </button>
        </div>
      ) : null}
    </section>
  );
}
