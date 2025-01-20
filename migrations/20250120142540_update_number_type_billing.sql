-- +goose Up
-- +goose StatementBegin
ALTER TABLE billing
ALTER COLUMN number TYPE TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE billing
ALTER COLUMN number TYPE INT;
-- +goose StatementEnd
