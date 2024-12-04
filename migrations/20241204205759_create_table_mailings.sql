-- +goose Up
-- +goose StatementBegin
CREATE TABLE mailings (
  id SERIAL PRIMARY KEY, 
  type_id INT NOT NULL REFERENCES mailing_types(id),
  markdown TEXT NOT NULL,
  html TEXT,
  sent BOOL,
  sent_on TIMESTAMPTZ,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE mailings;
-- +goose StatementEnd
