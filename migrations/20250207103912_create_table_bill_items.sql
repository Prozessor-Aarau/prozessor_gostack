-- +goose Up
-- +goose StatementBegin
CREATE TABLE bill_items (
  id UUID PRIMARY KEY,
  bill_id INT NOT NULL REFERENCES billing(id) ON DELETE CASCADE,
  position INT NOT NULL,
  description TEXT NOT NULL,
  unit_price REAL NOT NULL,
  count REAL NOT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE bill_items;
-- +goose StatementEnd
