import type { Template } from "../types";

// TemplatesView renders shared templates and apply actions.
export function TemplatesView({
  templates,
  loading,
  currentUserName,
  authHeaderEnabled,
  hasUser,
  onRefresh,
  onApplyTemplate,
}: {
  templates: Template[];
  loading: boolean;
  currentUserName: string;
  authHeaderEnabled: boolean;
  hasUser: boolean;
  onRefresh: () => void;
  onApplyTemplate: (templateId: string) => void;
}) {
  return (
    <section className="panel">
      <div className="panel-header">
        <div>
          <h3>Templates</h3>
          <p className="muted small">Create workouts from shared templates.</p>
        </div>
        <div className="actions">
          {authHeaderEnabled && (
            <div className="muted small">{currentUserName}</div>
          )}
          <button className="btn subtle" onClick={onRefresh}>
            Refresh
          </button>
        </div>
      </div>
      {loading && <p>Loading templates…</p>}
      {/* Template list */}
      <div className="list">
        {templates.map((tmpl) => (
          <div key={tmpl.id} className="list-item">
            <div className="list-row">
              <div>
                <strong>{tmpl.name}</strong>
                <div className="muted small">
                  {tmpl.steps?.length || 0} steps
                </div>
              </div>
              <div className="btn-group">
                <button
                  className="btn primary"
                  disabled={!hasUser}
                  onClick={() => onApplyTemplate(tmpl.id)}
                >
                  Use for user
                </button>
              </div>
            </div>
          </div>
        ))}
        {!loading && !templates.length && (
          <p className="muted">No templates yet.</p>
        )}
      </div>
    </section>
  );
}
