-- +goose Up
-- +goose StatementBegin
CREATE TABLE visitor_experts (
  id SERIAL PRIMARY KEY, 
  visitor_id UUID REFERENCES visitors(id) ON DELETE CASCADE, 
  atelier_id INT REFERENCES ateliers(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE visitor_experts;
-- +goose StatementEnd
