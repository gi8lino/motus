import { useState } from "react";

import type {
  AskConfirmOptions,
  CatalogExercise,
  SoundOption,
  Workout,
} from "../../types";

import { WorkoutsList } from "../workouts/WorkoutList";
import { WorkoutsEditor } from "../workouts/WorkoutEditor";
import { useWorkoutActions } from "../../hooks/useWorkoutActions";

export type WorkoutsViewProps = {
  workouts: Workout[] | null;
  loading?: boolean;

  setWorkouts: (updater: (prev: Workout[] | null) => Workout[] | null) => void;

  currentUserId: string | null;

  // Form dependencies (required by WorkoutForm)
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

  // Dialog helpers
  askConfirm: (
    message: string,
    options?: AskConfirmOptions,
  ) => Promise<boolean>;
  askPrompt: (message: string, defaultValue?: string) => Promise<string | null>;

  // Templates
  templatesReload: () => void;

  onToast?: (message: string) => void;
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
      askConfirm: props.askConfirm,
      askPrompt: props.askPrompt,
      notify: props.notifyUser,
      templatesReload: props.templatesReload,
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
        sounds={props.sounds}
        exerciseCatalog={props.exerciseCatalog}
        onCreateExercise={props.onCreateExercise}
        promptUser={props.promptUser}
        notifyUser={props.notifyUser}
        defaults={{
          defaultStepSoundKey: props.defaultStepSoundKey,
          defaultPauseDuration: props.defaultPauseDuration,
          defaultPauseSoundKey: props.defaultPauseSoundKey,
          defaultPauseAutoAdvance: props.defaultPauseAutoAdvance,
          repeatRestAfterLastDefault: props.repeatRestAfterLastDefault,
        }}
        askConfirm={(msg) => props.askConfirm(msg)}
        onToast={props.onToast}
      />
    </>
  );
}
