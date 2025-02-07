-- +goose Up
-- +goose StatementBegin
ALTER TABLE billing
ADD COLUMN custom BOOL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE billing
DROP COLUMN custom;
-- +goose StatementEnd
