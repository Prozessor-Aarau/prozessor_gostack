-- +goose Up
-- +goose StatementBegin
ALTER TABLE roles
    ADD COLUMN edit_opening_hours bool DEFAULT FALSE NOT NULL,
    ADD COLUMN delete_opening_hours bool DEFAULT FALSE NOT NULL,
    ADD COLUMN view_supervisions bool DEFAULT FALSE NOT NULL,
    ADD COLUMN edit_own_supervisions bool DEFAULT FALSE NOT NULL,
    ADD COLUMN edit_all_supervisions bool DEFAULT FALSE NOT NULL;

UPDATE
    roles
SET
    edit_opening_hours = TRUE,
    delete_opening_hours = TRUE,
    view_supervisions = TRUE,
    edit_own_supervisions = TRUE,
    edit_all_supervisions = TRUE
WHERE
    role_name = 'Superadmin';

UPDATE
    roles
SET
    edit_opening_hours = FALSE,
    delete_opening_hours = FALSE,
    view_supervisions = TRUE,
    edit_own_supervisions = TRUE,
    edit_all_supervisions = TRUE
WHERE
    role_name = 'Admin'
    OR role_name = 'Finanzen'
    OR role_name = 'Vorstandsmitglied'
    OR role_name = 'Ressort Leitung';

UPDATE
    roles
SET
    edit_opening_hours = FALSE,
    delete_opening_hours = FALSE,
    view_supervisions = TRUE,
    edit_own_supervisions = TRUE,
    edit_all_supervisions = FALSE
WHERE
    role_name = 'Ressort Mitglied';

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE roles
    DROP COLUMN edit_opening_hours,
    DROP COLUMN delete_opening_hours,
    DROP COLUMN view_supervisions,
    DROP COLUMN edit_own_supervisions,
    DROP COLUMN edit_all_supervisions;

-- +goose StatementEnd
