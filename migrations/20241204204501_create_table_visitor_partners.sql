-- +goose Up
-- +goose StatementBegin
CREATE TABLE visitor_partners (
  visitor_id UUID REFERENCES visitors(id) ON DELETE CASCADE, 
  partner_id INT REFERENCES partners(id) ON DELETE CASCADE,
  PRIMARY KEY (visitor_id, partner_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE visitor_partners;
-- +goose StatementEnd
