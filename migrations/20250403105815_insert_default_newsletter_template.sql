-- +goose Up
-- +goose StatementBegin
INSERT INTO newsletter_templates (name, description, template)
VALUES('Newsletter Basic', 'Basis Newsletter Template mit 2 dynamischen Farben und dem schraffierten Prozessor Logo im Titel-Banner', 'basic_newsletter.html');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM newsletter_templates;
-- +goose StatementEnd
