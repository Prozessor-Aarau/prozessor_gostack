-- +goose Up
-- +goose StatementBegin
ALTER TABLE logistics
ADD COLUMN unit_price real DEFAULT null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE logistics
DROP COLUMN unit_price;
-- +goose StatementEnd
