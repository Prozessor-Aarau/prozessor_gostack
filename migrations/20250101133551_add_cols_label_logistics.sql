-- +goose Up
-- +goose StatementBegin
ALTER TABLE logistics
ADD COLUMN label text DEFAULT null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE logistics
DROP COLUMN label;
-- +goose StatementEnd
