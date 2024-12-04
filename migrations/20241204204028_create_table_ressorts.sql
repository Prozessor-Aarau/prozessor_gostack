-- +goose Up
-- +goose StatementBegin
CREATE TABLE ressorts (
  id serial PRIMARY KEY,
  ressort_name TEXT NOT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE ressorts;
-- +goose StatementEnd
