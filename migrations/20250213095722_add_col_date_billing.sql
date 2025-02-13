-- +goose Up
-- +goose StatementBegin
ALTER TABLE billing
ADD COLUMN date TIMESTAMPTZ DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE billing
DROP COLUMN date;
-- +goose StatementEnd
