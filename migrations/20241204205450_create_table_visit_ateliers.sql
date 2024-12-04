-- +goose Up
-- +goose StatementBegin
CREATE TABLE visit_ateliers (
  visit_id INT NOT NULL REFERENCES visits(id) ON DELETE CASCADE,
  atelier_id INT NOT NULL REFERENCES ateliers(id),
  PRIMARY KEY (visit_id, atelier_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE visit_ateliers;
-- +goose StatementEnd
