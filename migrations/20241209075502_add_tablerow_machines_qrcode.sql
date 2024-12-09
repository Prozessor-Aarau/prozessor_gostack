-- +goose Up
-- +goose StatementBegin
ALTER TABLE machines
ADD COLUMN qr_code TEXT DEFAULT NULL;-- +goose StatementEnd
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE machines
DROP COLUMN qr_code;
-- +goose StatementEnd
