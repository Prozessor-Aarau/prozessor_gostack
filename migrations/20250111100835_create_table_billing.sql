-- +goose Up
-- +goose StatementBegin
CREATE TABLE billing (
  id SERIAL PRIMARY KEY,
  purpose TEXT NOT NULL,
  visitor_id UUID REFERENCES visitors(id) ON DELETE SET NULL,
  with_address bool DEFAULT false,
  first_name TEXT DEFAULT NULL,
  family_name TEXT DEFAULT NULL,
  street TEXT DEFAULT NULL,
  number INT DEFAULT NULL,
  postcode VARCHAR(5) DEFAULT NULL,
  city TEXT DEFAULT NULL,
  amount REAL NOT NULL,
  infos TEXT DEFAULT NULL,
  with_message bool DEFAULT false,
  message TEXT DEFAULT NULL,
  with_reference bool DEFAULT false,
  reference TEXT DEFAULT NULL,
  file_name TEXT DEFAULT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE billing;
-- +goose StatementEnd
