-- +goose Up
-- +goose StatementBegin
CREATE TABLE machine_faqs (
  id SERIAL PRIMARY KEY, 
  machine_id UUID NOT NULL references machines(id) ON DELETE CASCADE,
  question TEXT NOT NULL,
  answer TEXT NOT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE machine_faqs;
-- +goose StatementEnd
