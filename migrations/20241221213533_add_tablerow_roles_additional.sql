-- +goose Up
-- +goose StatementBegin
UPDATE roles SET role_name = 'Ressort Leitung' WHERE id = 3;
UPDATE roles SET role_name = 'Logistik' WHERE id = 4;

INSERT INTO roles (role_name)
VALUES ('Admin'),
      ('Vorstandsmitglied'),
      ('Finanzen'),
      ('Superadmin');

UPDATE visitors SET role_id = 8 WHERE role_id = 4;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE visitors set role_id = 4 WHERE role_id = 8;
DELETE FROM roles WHERE role_name = 'Superadmin';
DELETE FROM roles WHERE role_name = 'Finanzen';
DELETE FROM roles WHERE role_name = 'Vorstandsmitglied';
DELETE FROM roles WHERE role_name = 'Admin';

UPDATE roles SET role_name = 'Superadmin' WHERE id = 4;
UPDATE roles SET role_name = 'Vorstandsmitglied' WHERE id = 3;
ALTER SEQUENCE public.roles_id_seq RESTART WITH 5;
-- +goose StatementEnd
