-- +goose Up
-- +goose StatementBegin
UPDATE
    newsletter_elements
SET
    content = REPLACE(content, 'EventTime', 'StartTime');

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
UPDATE
    newsletter_elements
SET
    content = REPLACE(content, 'StartTime', 'EventTime');

-- +goose StatementEnd
