-- +goose Up
-- +goose StatementBegin
CREATE TABLE badges (
  id SERIAL PRIMARY KEY, 
  number INT NOT NULL UNIQUE,
  uid TEXT NOT NULL,
  visitor_id UUID REFERENCES visitors(id) ON DELETE SET NULL,
  extern_first_name TEXT,
  extern_family_name TEXT,
  extern_city TEXT,
  extern_phone TEXT,
  extern_email TEXT,
  extern_purpose TEXT,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE badges;
-- +goose StatementEnd
