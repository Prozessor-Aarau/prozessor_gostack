-- +goose Up
-- +goose StatementBegin
ALTER TABLE 
  log 
RENAME COLUMN 
  type TO level;

ALTER TABLE 
  log 
RENAME COLUMN 
  module TO location;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE
  log
RENAME COLUMN
  level TO
type;

ALTER TABLE
  log
RENAME COLUMN
  location TO module;
-- +goose StatementEnd
