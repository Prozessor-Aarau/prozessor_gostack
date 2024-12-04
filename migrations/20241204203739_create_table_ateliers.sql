-- +goose Up
-- +goose StatementBegin
CREATE TABLE ateliers (
  id serial PRIMARY KEY,
  atelier_name TEXT,
  description TEXT,
  active BOOL,
  icon TEXT,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE ateliers;
-- +goose StatementEnd
