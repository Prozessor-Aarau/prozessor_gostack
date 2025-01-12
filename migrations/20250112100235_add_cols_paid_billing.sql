-- +goose Up
-- +goose StatementBegin
ALTER TABLE billing
ADD COLUMN paid bool DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE billing
DROP COLUMN paid;
-- +goose StatementEnd
