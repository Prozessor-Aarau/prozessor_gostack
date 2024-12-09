-- +goose Up
-- +goose StatementBegin
ALTER TABLE machines
ADD COLUMN qr_codes TEXT DEFAULT NULL;-- +goose StatementEnd
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE machines
DROP COLUMN qr_codes;
-- +goose StatementEnd
