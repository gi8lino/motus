import { useState } from "react";

import type { SessionHistoryItem, SessionState, Workout } from "../../types";
import { HistoryList } from "./HistoryCard";
import { HistoryPreviewModal } from "./HistoryPreviewModal";

// HistoryView lists logged sessions and opens a session preview.
export function HistoryView({
  items,
  activeSession,
  onResume,
  loadWorkout,
  onCopySummary,
}: {
  items: SessionHistoryItem[];
  activeSession: SessionState | null;
  onResume: () => void;
  loadWorkout: (id: string) => Promise<Workout>;
  onCopySummary: () => void;
}) {
  const [preview, setPreview] = useState<SessionHistoryItem | null>(null);
  const [previewWorkout, setPreviewWorkout] = useState<Workout | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  // handleSelect loads preview details for a selected session.
  const handleSelect = (item: SessionHistoryItem) => {
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
          activeSession={activeSession}
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
