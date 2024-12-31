-- +goose Up
-- +goose StatementBegin
ALTER TABLE roles
ADD COLUMN edit_logistics BOOL DEFAULT false;-- +goose StatementEnd

UPDATE roles
SET edit_logistics = true
WHERE role_name = 'Superadmin' OR role_name = 'Logistik';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE roles
DROP COLUMN edit_logistics;
-- +goose StatementEnd
