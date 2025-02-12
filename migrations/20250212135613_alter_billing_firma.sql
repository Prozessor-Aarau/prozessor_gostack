-- +goose Up
-- +goose StatementBegin
ALTER TABLE billing
ADD COLUMN firma TEXT DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE billing
DROP COLUMN firma;
-- +goose StatementEnd
