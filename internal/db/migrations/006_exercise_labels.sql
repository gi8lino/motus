CREATE TABLE IF NOT EXISTS exercise_labels (
  exercise_id TEXT NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
  label TEXT NOT NULL,
  PRIMARY KEY (exercise_id, label),
  CHECK (label = LOWER(TRIM(label)) AND label <> '')
);

CREATE INDEX IF NOT EXISTS exercise_labels_label_idx ON exercise_labels(label);
