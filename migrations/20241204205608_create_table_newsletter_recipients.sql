-- +goose Up
-- +goose StatementBegin
CREATE TABLE newsletter_recipients (
  id SERIAL PRIMARY KEY, 
  email TEXT UNIQUE NOT NULL,
  first_name TEXT NOT NULL,
  family_name TEXT NOT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE newsletter_recipients;
-- +goose StatementEnd
