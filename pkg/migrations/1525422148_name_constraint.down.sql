ALTER TABLE solutions
  ALTER COLUMN name SET NOT NULL,
  DROP CONSTRAINT "name_user_id";