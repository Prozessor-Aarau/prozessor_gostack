-- +goose Up
-- +goose StatementBegin
ALTER TABLE newsletters
ADD COLUMN color_base VARCHAR(10) NOT NULL,
ADD COLUMN color_accent VARCHAR(10) NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE newsletters
DROP COLUMN color_base,
DROP COLUMN color_accent;
-- +goose StatementEnd
