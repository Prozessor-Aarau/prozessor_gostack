-- +goose Up
-- +goose StatementBegin
CREATE TABLE abonnements (
  id serial PRIMARY KEY,
  buydate DATE NOT NULL,
  partial_payment BOOL DEFAULT false NOT NULL,
  visitor_id UUID REFERENCES visitors(id) ON DELETE SET NULL,
  reminder TIMESTAMPTZ,
  member BOOL DEFAULT false NOT NULL,
  price INT NOT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE abonnements;
-- +goose StatementEnd
