-- +goose Up
-- +goose StatementBegin
ALTER TABLE logistics
    ADD COLUMN request_pending bool DEFAULT FALSE;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE logistics
    DROP COLUMN request_pending;

-- +goose StatementEnd
