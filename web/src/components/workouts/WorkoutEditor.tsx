import { useState } from "react";
import type { CatalogExercise, SoundOption, Workout } from "../../types";
import { WorkoutForm, type WorkoutFormDefaults } from "./WorkoutForm";
import { useWorkoutFormActions } from "../../hooks/useWorkoutFormActions";
import { Modal } from "../common/Modal";

type WorkoutFormData = {
  sounds: SoundOption[];
  exerciseCatalog: CatalogExercise[];
};

type WorkoutFormServices = {
  onCreateExercise: (name: string) => Promise<CatalogExercise>;
  promptUser: (
    message: string,
    defaultValue?: string,
  ) => Promise<string | null>;
  notifyUser: (message: string) => Promise<void>;
  askConfirm: (message: string) => Promise<boolean>;
  onToast?: (message: string) => void;
};

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
  formData: WorkoutFormData;
  defaults: WorkoutFormDefaults;
  services: WorkoutFormServices;
};

export function WorkoutsEditor({
  open,
  onClose,
  editingWorkout,
  setEditingWorkout,
  currentUserId,
  setWorkouts,
  setSelectedWorkoutId,
  formData,
  defaults,
  services,
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
      askConfirm: services.askConfirm,
    });

  return (
    <Modal
      open={open}
      onClose={() => {
        closeWorkoutModal();
        onClose();
      }}
    >
      <WorkoutForm
        userId={currentUserId}
        sounds={formData.sounds}
        exerciseCatalog={formData.exerciseCatalog}
        defaults={defaults}
        services={{
          onSave: saveWorkout,
          onUpdate: updateWorkout,
          onCreateExercise: services.onCreateExercise,
          promptUser: services.promptUser,
          notifyUser: services.notifyUser,
          onToast: services.onToast,
        }}
        editingWorkout={editingWorkout}
        onDirtyChange={setWorkoutDirty}
        onClose={() => {
          closeWorkoutModal();
          onClose();
        }}
      />
    </Modal>
  );
}
