DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'workout_exercise_side_check'
      AND conrelid = 'workout_subset_exercises'::regclass
  ) THEN
    ALTER TABLE workout_subset_exercises
      ADD CONSTRAINT workout_exercise_side_check
      CHECK (side IN ('left', 'right', 'not_applicable'));
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'workout_exercise_type_check'
      AND conrelid = 'workout_subset_exercises'::regclass
  ) THEN
    ALTER TABLE workout_subset_exercises
      ADD CONSTRAINT workout_exercise_type_check
      CHECK (exercise_type IN ('rep', 'stopwatch', 'countdown'));
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'workout_step_estimate_check'
      AND conrelid = 'workout_steps'::regclass
  ) THEN
    ALTER TABLE workout_steps
      ADD CONSTRAINT workout_step_estimate_check
      CHECK (estimated_seconds >= 0);
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'workout_subset_estimate_check'
      AND conrelid = 'workout_subsets'::regclass
  ) THEN
    ALTER TABLE workout_subsets
      ADD CONSTRAINT workout_subset_estimate_check
      CHECK (estimated_seconds >= 0);
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'training_step_elapsed_check'
      AND conrelid = 'training_steps'::regclass
  ) THEN
    ALTER TABLE training_steps
      ADD CONSTRAINT training_step_elapsed_check
      CHECK (elapsed_millis >= 0);
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS workouts_user_id_idx ON workouts(user_id);
CREATE INDEX IF NOT EXISTS workout_trainings_user_completed_idx ON workout_trainings(user_id, completed_at DESC);
CREATE INDEX IF NOT EXISTS workout_steps_workout_order_idx ON workout_steps(workout_id, step_order);
CREATE INDEX IF NOT EXISTS workout_subsets_step_order_idx ON workout_subsets(step_id, subset_order);
CREATE INDEX IF NOT EXISTS workout_exercises_subset_order_idx ON workout_subset_exercises(subset_id, exercise_order);
CREATE INDEX IF NOT EXISTS exercises_lower_name_idx ON exercises(lower(name));
