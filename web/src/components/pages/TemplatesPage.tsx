import type { Template } from "../../types";
import { UI_TEXT } from "../../utils/uiText";

export type TemplatesViewData = {
  templates: Template[];
  loading: boolean;
  hasUser: boolean;
};

export type TemplatesViewActions = {
  onRefresh: () => void;
  onApplyTemplate: (templateId: string) => void;
};

// TemplatesView renders shared templates and apply actions.
export function TemplatesView({
  data,
  actions,
}: {
  data: TemplatesViewData;
  actions: TemplatesViewActions;
}) {
  const { templates, loading, hasUser } = data;
  const { onRefresh, onApplyTemplate } = actions;
  return (
    <section className="panel">
      <div className="panel-header">
        <div>
          <h3>{UI_TEXT.pages.templates.title}</h3>
          <p className="muted small">{UI_TEXT.pages.templates.hint}</p>
        </div>
        <div className="actions">
          <button className="btn subtle" onClick={onRefresh}>
            {UI_TEXT.pages.templates.refresh}
          </button>
        </div>
      </div>
      {loading && <p>{UI_TEXT.pages.templates.loading}</p>}
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
                  {UI_TEXT.pages.templates.apply}
                </button>
              </div>
            </div>
          </div>
        ))}
        {!loading && !templates.length && (
          <p className="muted">{UI_TEXT.pages.templates.empty}</p>
        )}
      </div>
    </section>
  );
}
