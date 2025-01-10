-- +goose Up
-- +goose StatementBegin
ALTER TABLE logistics
ADD COLUMN paid bool DEFAULT null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE logistics
DROP COLUMN paid;
-- +goose StatementEnd
