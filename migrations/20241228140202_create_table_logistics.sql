-- +goose Up
-- +goose StatementBegin
CREATE TABLE logistics (
  id serial PRIMARY KEY,
  number INT NOT NULL,
  start_date DATE,
  end_date DATE,
  category INT REFERENCES logistics_categories(id) ON DELETE CASCADE NOT NULL,
  visitor_id UUID REFERENCES visitors(id) ON DELETE SET NULL,
  reminder TIMESTAMPTZ,
  member BOOL DEFAULT false NOT NULL,
  price REAL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE logistics;
-- +goose StatementEnd
