import { useState } from "react";

import type { TrainngHistoryItem, TrainngState, Workout } from "../../types";
import { HistoryList } from "../history/HistoryCard";
import { HistoryPreviewModal } from "../history/HistoryPreviewModal";

// HistoryView lists logged trainings and opens a training preview.
export function HistoryView({
  items,
  activeTraining,
  onResume,
  loadWorkout,
  onCopySummary,
}: {
  items: TrainngHistoryItem[];
  activeTraining: TrainngState | null;
  onResume: () => void;
  loadWorkout: (id: string) => Promise<Workout>;
  onCopySummary: () => void;
}) {
  const [preview, setPreview] = useState<TrainngHistoryItem | null>(null);
  const [previewWorkout, setPreviewWorkout] = useState<Workout | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  // handleSelect loads preview details for a selected training.
  const handleSelect = (item: TrainngHistoryItem) => {
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
            <h3>Training history</h3>
            <p className="muted small">
              Completed trainings for the selected user.
            </p>
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
