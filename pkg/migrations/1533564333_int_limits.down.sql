DROP FUNCTION IF EXISTS pc_chartoint;

ALTER TABLE templates
  ALTER COLUMN cpu TYPE TEXT,
  ALTER COLUMN ram TYPE TEXT;
