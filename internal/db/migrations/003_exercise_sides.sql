ALTER TABLE workout_subset_exercises ADD COLUMN IF NOT EXISTS side TEXT NOT NULL DEFAULT 'not_applicable';
ALTER TABLE exercises ADD COLUMN IF NOT EXISTS has_sides BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE workout_subset_exercises SET side = CASE WHEN name ~* '[[:space:]]left$' THEN 'left' WHEN name ~* '[[:space:]]right$' THEN 'right' ELSE side END, name = regexp_replace(name, '[[:space:]]+(left|right)$', '', 'i') WHERE name ~* '[[:space:]](left|right)$';
DO $$
DECLARE duplicate RECORD;
DECLARE target_id TEXT;
BEGIN
  FOR duplicate IN SELECT id, regexp_replace(name, '[[:space:]]+(left|right)$', '', 'i') AS base_name FROM exercises WHERE name ~* '[[:space:]](left|right)$' LOOP
    SELECT id INTO target_id FROM exercises WHERE lower(name) = lower(duplicate.base_name) AND id <> duplicate.id LIMIT 1;
    IF target_id IS NULL THEN
      UPDATE exercises SET name = duplicate.base_name, has_sides = TRUE WHERE id = duplicate.id;
      target_id := duplicate.id;
    ELSE
      UPDATE exercises SET has_sides = TRUE WHERE id = target_id;
      UPDATE workout_subset_exercises SET exercise_id = target_id WHERE exercise_id = duplicate.id;
      DELETE FROM exercises WHERE id = duplicate.id;
    END IF;
    target_id := NULL;
  END LOOP;
END $$;
UPDATE workout_subset_exercises wse SET name = e.name FROM exercises e WHERE wse.exercise_id = e.id AND e.has_sides = TRUE;
