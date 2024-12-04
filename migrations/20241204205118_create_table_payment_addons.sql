-- +goose Up
-- +goose StatementBegin
CREATE TABLE payment_addons (
  id SERIAL PRIMARY KEY, 
  description TEXT,
  price REAL,
  payment_id INT REFERENCES payments(id),
  atelier_id INT REFERENCES ateliers(id) DEFAULT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE payment_addons;
-- +goose StatementEnd
