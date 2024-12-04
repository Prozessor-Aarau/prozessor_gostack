-- +goose Up
-- +goose StatementBegin
INSERT INTO mailing_types (name, description)
VALUES ('Internes Mailing', 'Mailing, welches an alle aktiven Vereinsmitglieder gesendet werden soll'),
      ('Abo Reminder', 'Reminder an die Abonnent:innen, wenn ein Abo ablaufen könnte'),
      ('Newsletter', 'Newsletter an alle, die sich für den Newsletter angemeldet haben');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM mailing_types;
-- +goose StatementEnd
