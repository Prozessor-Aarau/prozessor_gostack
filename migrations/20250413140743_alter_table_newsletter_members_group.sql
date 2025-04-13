-- +goose Up
-- +goose StatementBegin
ALTER TABLE newsletters 
ADD COLUMN recipient_group TEXT;

UPDATE newsletters
SET recipient_group = 
CASE 
  WHEN members = true THEN 'mitglied'
  ELSE 'newsletter'
END;

ALTER TABLE newsletters
DROP COLUMN members;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE newsletters
ADD COLUMN members BOOL;

UPDATE newsletters
SET members = 
CASE 
  WHEN recipient_group = 'mitglied' THEN true
  ELSE false
END;

ALTER TABLE newsletters
DROP COLUMN recipient_group;
-- +goose StatementEnd
