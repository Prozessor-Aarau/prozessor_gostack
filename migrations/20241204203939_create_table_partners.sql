-- +goose Up
-- +goose StatementBegin
CREATE TABLE partners (
  id serial PRIMARY KEY,
  partner_name TEXT NOT NULL,
  contact_firstname TEXT NOT NULL,
  contact_familyname TEXT NOT NULL,
  email TEXT NOT NULL,
  address TEXT,
  postcode VARCHAR(5) NOT NULL,
  city TEXT NOT NULL,
  description TEXT,
  active BOOL DEFAULT true NOT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE partners;
-- +goose StatementEnd
