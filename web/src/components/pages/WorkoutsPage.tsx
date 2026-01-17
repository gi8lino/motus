import { useState } from "react";

import type {
  AskConfirmOptions,
  CatalogExercise,
  SoundOption,
  Workout,
} from "../../types";
import type { WorkoutFormDefaults } from "../workouts/WorkoutForm";

import { WorkoutsList } from "../workouts/WorkoutList";
import { WorkoutsEditor } from "../workouts/WorkoutEditor";
import { useWorkoutActions } from "../../hooks/useWorkoutActions";

export type WorkoutsServices = {
  onCreateExercise: (name: string) => Promise<CatalogExercise>;
  promptUser: (
    message: string,
    defaultValue?: string,
  ) => Promise<string | null>;
  notifyUser: (message: string) => Promise<void>;
  askConfirm: (
    message: string,
    options?: AskConfirmOptions,
  ) => Promise<boolean>;
  askPrompt: (message: string, defaultValue?: string) => Promise<string | null>;
  templatesReload: () => void;
  onToast?: (message: string) => void;
};

type WorkoutFormData = {
  sounds: SoundOption[];
  exerciseCatalog: CatalogExercise[];
};

export type WorkoutsViewProps = {
  workouts: Workout[] | null;
  loading?: boolean;

  setWorkouts: (updater: (prev: Workout[] | null) => Workout[] | null) => void;

  currentUserId: string | null;

  defaults: WorkoutFormDefaults;
  formData: WorkoutFormData;
  services: WorkoutsServices;
};

export function WorkoutsView(props: WorkoutsViewProps) {
  const workouts = props.workouts || [];

  const [selectedWorkoutId, setSelectedWorkoutId] = useState<string | null>(
    null,
  );
  const [editingWorkout, setEditingWorkout] = useState<Workout | null>(null);
  const [editorOpen, setEditorOpen] = useState(false);

  // Keep using existing action logic for "new/edit" selection behavior
  const { newWorkout, editWorkoutFromList, removeWorkout, shareWorkout } =
    useWorkoutActions({
      workouts,
      selectedWorkoutId,
      setEditingWorkout,
      setSelectedWorkoutId,
      setWorkouts: props.setWorkouts,
      askConfirm: props.services.askConfirm,
      askPrompt: props.services.askPrompt,
      notify: props.services.notifyUser,
      templatesReload: props.services.templatesReload,
    });

  return (
    <>
      <WorkoutsList
        workouts={workouts}
        loading={props.loading}
        currentUserId={props.currentUserId}
        setSelectedWorkoutId={setSelectedWorkoutId}
        onNew={() => newWorkout()}
        onEdit={(id) => editWorkoutFromList(id)}
        onOpenEditor={() => setEditorOpen(true)}
        onShare={(id) => shareWorkout(id)}
        onDelete={(id) => removeWorkout(id)}
      />

      <WorkoutsEditor
        open={editorOpen}
        onClose={() => setEditorOpen(false)}
        editingWorkout={editingWorkout}
        setEditingWorkout={setEditingWorkout}
        currentUserId={props.currentUserId}
        setWorkouts={props.setWorkouts}
        setSelectedWorkoutId={setSelectedWorkoutId}
        formData={props.formData}
        defaults={props.defaults}
        services={{
          onCreateExercise: props.services.onCreateExercise,
          promptUser: props.services.promptUser,
          notifyUser: props.services.notifyUser,
          askConfirm: props.services.askConfirm,
          onToast: props.services.onToast,
        }}
      />
    </>
  );
}
