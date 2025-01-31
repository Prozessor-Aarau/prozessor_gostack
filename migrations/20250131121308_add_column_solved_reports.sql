-- +goose Up
-- +goose StatementBegin
ALTER TABLE reports
ADD COLUMN solved bool DEFAULT false NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE reports
DROP COLUMN solved;
-- +goose StatementEnd
