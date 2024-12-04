-- +goose Up
-- +goose StatementBegin
CREATE TABLE payments (
  id SERIAL PRIMARY KEY, 
  baseprice REAL,
  paidby TEXT,
  visit_id INT NOT NULL REFERENCES visits(id),
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE payments;
-- +goose StatementEnd
