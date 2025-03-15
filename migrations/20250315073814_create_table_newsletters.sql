-- +goose Up
-- +goose StatementBegin
CREATE TABLE newsletter_templates (
  id serial PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  template TEXT NOT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE newsletters (
  id serial PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  template_id INT REFERENCES newsletter_templates(id) ON DELETE CASCADE NOT NULL,
  members BOOL DEFAULT false,
  sent_count INT,
  sent_on DATE,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE newsletter_elements (
  id serial PRIMARY KEY,
  newsletter_id INT REFERENCES newsletters(id) ON DELETE CASCADE NOT NULL,
  position INT NOT NULL,
  type VARCHAR(100) NOT NULL,
  content TEXT NOT NULL,
  upload_url TEXT DEFAULT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE newsletter_elements;
DROP TABLE newsletters;
DROP TABLE newsletter_templates;
-- +goose StatementEnd
