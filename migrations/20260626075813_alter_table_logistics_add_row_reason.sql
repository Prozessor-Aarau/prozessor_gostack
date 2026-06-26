-- +goose Up
-- +goose StatementBegin
ALTER TABLE logistics
    ADD COLUMN reason text;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE logistics
    DROP COLUMN reason;

-- +goose StatementEnd
