-- +goose Up
-- +goose StatementBegin
ALTER TABLE ressorts
ADD COLUMN slack_channel TEXT DEFAULT NULL;-- +goose StatementEnd
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE ressorts
DROP COLUMN slack_channel;
-- +goose StatementEnd
