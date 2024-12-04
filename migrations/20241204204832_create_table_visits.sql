-- +goose Up
-- +goose StatementBegin
CREATE TABLE visits (
  id SERIAL PRIMARY KEY, 
  abonnement BOOL,
  date_in TIMESTAMPTZ, 
  date_out TIMESTAMPTZ,
  partner_id INT REFERENCES partners(id) ON DELETE SET NULL,
  visitor_id UUID REFERENCES visitors(id) ON DELETE SET NULL,
  ressort_id INT REFERENCES ressorts(id) ON DELETE SET NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE visits;
-- +goose StatementEnd
