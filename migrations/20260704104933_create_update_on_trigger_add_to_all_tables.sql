-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_on()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_on = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
DECLARE
  t TEXT;
BEGIN
  FOR t IN
    SELECT table_name
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND column_name = 'updated_on'
  LOOP
    EXECUTE format('
      CREATE OR REPLACE TRIGGER trg_set_updated_on
      BEFORE UPDATE ON %I
      FOR EACH ROW EXECUTE FUNCTION set_updated_on();
    ', t);
  END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
  t TEXT;
BEGIN
  FOR t IN
    SELECT table_name
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND column_name = 'updated_on'
  LOOP
    EXECUTE format('DROP TRIGGER IF EXISTS trg_set_updated_on ON %I;', t);
  END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS set_updated_on();
-- +goose StatementEnd
