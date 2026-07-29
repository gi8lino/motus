ALTER TABLE workout_trainings ADD COLUMN notes TEXT NOT NULL DEFAULT '';
ALTER TABLE workout_trainings ADD COLUMN perceived_effort SMALLINT;
ALTER TABLE workout_trainings ADD CONSTRAINT workout_trainings_effort_check
  CHECK (perceived_effort IS NULL OR perceived_effort BETWEEN 1 AND 10);
