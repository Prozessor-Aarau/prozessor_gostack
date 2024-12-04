-- +goose Up
-- +goose StatementBegin
CREATE TABLE mailing_types (
  id SERIAL PRIMARY KEY, 
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE mailing_types;
-- +goose StatementEnd
