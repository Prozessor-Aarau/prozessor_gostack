-- +goose Up
-- +goose StatementBegin
ALTER TABLE machines
    ADD COLUMN location text DEFAULT NULL;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE machines
    DROP COLUMN location;

-- +goose StatementEnd
