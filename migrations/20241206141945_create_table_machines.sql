-- +goose Up
-- +goose StatementBegin
CREATE TABLE machines (
  id SERIAL PRIMARY KEY, 
  ressort_id INT NOT NULL references ressorts(id) ON DELETE CASCADE,
  machine_name TEXT NOT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE machines;
-- +goose StatementEnd
