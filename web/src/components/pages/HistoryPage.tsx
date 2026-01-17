import { useState } from "react";

import type { TrainingHistoryItem, TrainingState, Workout } from "../../types";
import { HistoryList } from "../history/HistoryCard";
import { HistoryPreviewModal } from "../history/HistoryPreviewModal";
import { UI_TEXT } from "../../utils/uiText";

export type HistoryViewData = {
  items: TrainingHistoryItem[];
  activeTraining: TrainingState | null;
};

export type HistoryViewActions = {
  onResume: () => void;
  loadWorkout: (id: string) => Promise<Workout>;
  onCopySummary: () => void;
};

// HistoryView lists logged trainings and opens a training preview.
export function HistoryView({
  data,
  actions,
}: {
  data: HistoryViewData;
  actions: HistoryViewActions;
}) {
  const { items, activeTraining } = data;
  const { onResume, loadWorkout, onCopySummary } = actions;
  const [preview, setPreview] = useState<TrainingHistoryItem | null>(null);
  const [previewWorkout, setPreviewWorkout] = useState<Workout | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  // handleSelect loads preview details for a selected training.
  const handleSelect = (item: TrainingHistoryItem) => {
    setPreview(item);
    setPreviewLoading(true);
    loadWorkout(item.workoutId)
      .then((workout) => setPreviewWorkout(workout))
      .catch(() => setPreviewWorkout(null))
      .finally(() => setPreviewLoading(false));
  };

  return (
    <>
      <section className="panel">
        <div className="panel-header">
          <div>
            <h3>{UI_TEXT.pages.history.title}</h3>
            <p className="muted small">{UI_TEXT.pages.history.hint}</p>
          </div>
        </div>
        <HistoryList
          items={items}
          activeTraining={activeTraining}
          onResume={onResume}
          onSelect={handleSelect}
        />
      </section>
      {/* Preview modal */}
      <HistoryPreviewModal
        preview={preview}
        workout={previewWorkout}
        loading={previewLoading}
        onCopySummary={onCopySummary}
        onClose={() => {
          setPreview(null);
          setPreviewWorkout(null);
        }}
      />
    </>
  );
}
