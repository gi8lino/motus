import { useState } from "react";
import type { CatalogExercise, SoundOption, Workout } from "../../types";
import { WorkoutForm } from "./WorkoutForm";
import { useWorkoutFormActions } from "../../hooks/useWorkoutFormActions";

export type WorkoutsEditorProps = {
  open: boolean;
  onClose: () => void;

  editingWorkout: Workout | null;
  setEditingWorkout: (w: Workout | null) => void;

  currentUserId: string | null;

  // data
  setWorkouts: (updater: (prev: Workout[] | null) => Workout[] | null) => void;
  setSelectedWorkoutId: (id: string | null) => void;

  // form deps
  sounds: SoundOption[];
  exerciseCatalog: CatalogExercise[];
  onCreateExercise: (name: string) => Promise<CatalogExercise>;
  promptUser: (
    message: string,
    defaultValue?: string,
  ) => Promise<string | null>;
  notifyUser: (message: string) => Promise<void>;

  defaultStepSoundKey: string;
  defaultPauseDuration: string;
  defaultPauseSoundKey: string;
  defaultPauseAutoAdvance: boolean;
  repeatRestAfterLastDefault: boolean;

  askConfirm: (message: string) => Promise<boolean>;
  onToast?: (message: string) => void;
};

export function WorkoutsEditor({
  open,
  onClose,
  editingWorkout,
  setEditingWorkout,
  currentUserId,
  setWorkouts,
  setSelectedWorkoutId,
  sounds,
  exerciseCatalog,
  onCreateExercise,
  promptUser,
  notifyUser,
  defaultStepSoundKey,
  defaultPauseDuration,
  defaultPauseSoundKey,
  defaultPauseAutoAdvance,
  repeatRestAfterLastDefault,
  askConfirm,
  onToast,
}: WorkoutsEditorProps) {
  const [workoutDirty, setWorkoutDirty] = useState(false);

  const { saveWorkout, updateWorkout, closeWorkoutModal } =
    useWorkoutFormActions({
      currentUserId,
      workoutDirty,
      setSelectedWorkoutId,
      setEditingWorkout,
      setWorkoutDirty,
      setShowWorkoutForm: (show) => {
        // WorkoutsEditor controls open/close externally; mirror close only.
        if (!show) onClose();
      },
      setWorkouts,
      askConfirm,
    });

  return (
    <>
      {open ? (
        <div
          className="modal-overlay"
          onClick={() => {
            closeWorkoutModal();
            onClose();
          }}
        >
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <WorkoutForm
              userId={currentUserId}
              sounds={sounds}
              exerciseCatalog={exerciseCatalog}
              onCreateExercise={onCreateExercise}
              promptUser={promptUser}
              notifyUser={notifyUser}
              defaultStepSoundKey={defaultStepSoundKey}
              defaultPauseDuration={defaultPauseDuration}
              defaultPauseSoundKey={defaultPauseSoundKey}
              defaultPauseAutoAdvance={defaultPauseAutoAdvance}
              repeatRestAfterLastDefault={repeatRestAfterLastDefault}
              onSave={saveWorkout}
              onUpdate={updateWorkout}
              editingWorkout={editingWorkout}
              onDirtyChange={setWorkoutDirty}
              onToast={onToast}
              onClose={() => {
                closeWorkoutModal();
                onClose();
              }}
            />
          </div>
        </div>
      ) : null}
    </>
  );
}
