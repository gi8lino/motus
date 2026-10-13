ALTER TABLE workout_trainings ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';
ALTER TABLE workout_trainings ADD COLUMN IF NOT EXISTS perceived_effort SMALLINT;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'workout_trainings_effort_check'
      AND conrelid = 'workout_trainings'::regclass
  ) THEN
    ALTER TABLE workout_trainings
      ADD CONSTRAINT workout_trainings_effort_check
      CHECK (perceived_effort IS NULL OR perceived_effort BETWEEN 1 AND 10);
  END IF;
END $$;
