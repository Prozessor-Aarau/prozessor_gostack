-- +goose Up
-- +goose StatementBegin
INSERT INTO ressorts (ressort_name, slack_channel)
VALUES ('Holzwerkstatt', 'C07KETNRGV6'),
      ('Druck/Foto', 'C07JP7ZPJ3X'),
      ('Büro/Aufenthaltsraum', 'C06HGHFBCDP'),
      ('Makerspace/Laser/CNC', 'C07JBHQ18MD'),
      ('Metallwerkstatt', 'C07JP81T9RB'),
      ('Soundstudio', 'C06HT2455UG'),
      ('Textilatelier', 'C07JBHT9L07'),
      ('Velowerkstatt', 'C07KEU1F8TS');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM ressorts;
-- +goose StatementEnd
