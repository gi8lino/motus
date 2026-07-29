ALTER TABLE training_steps ADD COLUMN exercises JSONB NOT NULL DEFAULT '[]'::jsonb;
