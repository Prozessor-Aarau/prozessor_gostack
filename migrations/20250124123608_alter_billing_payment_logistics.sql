-- +goose Up
-- +goose StatementBegin
ALTER TABLE logistics
DROP COLUMN paid,
ADD COLUMN bill_id INT REFERENCES billing(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE logistics
ADD COLUMN paid bool DEFAULT null,
DROP COLUMN bill_id;
-- +goose StatementEnd
