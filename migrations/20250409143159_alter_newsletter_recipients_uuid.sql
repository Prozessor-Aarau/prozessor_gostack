-- +goose Up
-- +goose StatementBegin
ALTER TABLE newsletter_recipients
ADD COLUMN uid UUID NOT NULL,
ADD PRIMARY KEY (uid),
DROP COLUMN id;

ALTER TABLE newsletter_recipients
RENAME COLUMN uid TO id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE newsletter_recipients
ADD COLUMN sid SERIAL,
ADD PRIMARY KEY (sid),
DROP COLUMN id;

ALTER TABLE newsletter_recipients
RENAME COLUMN sid TO id;
-- +goose StatementEnd
