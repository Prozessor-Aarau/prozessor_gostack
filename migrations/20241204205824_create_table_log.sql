-- +goose Up
-- +goose StatementBegin
CREATE TABLE log (
  id SERIAL PRIMARY KEY, 
  type TEXT NOT NULL,
  module TEXT NOT NULL,
  message TEXT NOT NULL,
  visitor_id UUID REFERENCES visitors(id) ON DELETE SET NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE log;
-- +goose StatementEnd
