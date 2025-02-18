-- +goose Up
-- +goose StatementBegin
UPDATE billing SET date = oldbilling.created_on
FROM(
SELECT id, created_on
FROM billing 
WHERE date IS NULL) as oldbilling
WHERE oldbilling.id = billing.id;

ALTER TABLE billing
ALTER COLUMN date SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE billing
ALTER COLUMN date DROP NOT NULL;

UPDATE billing SET date = NULL
WHERE created_on = date;
-- +goose StatementEnd
