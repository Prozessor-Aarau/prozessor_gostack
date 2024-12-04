-- +goose Up
-- +goose StatementBegin
INSERT INTO roles (role_name)
VALUES ('Registriert'),
      ('Ressort Mitglied'),
      ('Vorstandsmitglied'),
      ('Superadmin');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM roles;
-- +goose StatementEnd
