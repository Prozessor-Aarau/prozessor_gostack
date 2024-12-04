-- +goose Up
-- +goose StatementBegin
CREATE TABLE visitors (
  id UUID PRIMARY KEY,
  first_name TEXT NOT NULL,
  family_name TEXT NOT NULL,
  birthdate DATE NOT NULL,
  address TEXT NOT NULL,
  postcode VARCHAR(5) NOT NULL,
  city TEXT NOT NULL,
  email TEXT UNIQUE,
  phone TEXT,
  infos TEXT,
  password TEXT,
  avatar VARCHAR(200),
  ressort_id INT REFERENCES ressorts(id) ON DELETE SET NULL,
  role_id INT REFERENCES roles(id) ON DELETE SET NULL,
  newsletter BOOL DEFAULT true NOT NULL,
  register_code UUID,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE visitors;
-- +goose StatementEnd
