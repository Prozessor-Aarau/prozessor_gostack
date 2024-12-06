-- +goose Up
-- +goose StatementBegin
CREATE TABLE reports (
  id SERIAL PRIMARY KEY, 
  ressort_id INT NOT NULL references ressorts(id) ON DELETE CASCADE,
  machine_id INT NOT NULL references machines(id) ON DELETE CASCADE,
  fatal bool NOT NULL DEFAULT false,
  report_info TEXT NOT NULL,
  report_reporter TEXT DEFAULT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE reports;
-- +goose StatementEnd
