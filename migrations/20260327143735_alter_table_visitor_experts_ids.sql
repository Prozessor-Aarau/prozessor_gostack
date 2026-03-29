-- +goose Up
-- +goose StatementBegin
ALTER TABLE visitor_experts
    DROP CONSTRAINT visitor_experts_pkey;

ALTER TABLE visitor_experts
    DROP COLUMN id;

ALTER TABLE visitor_experts
    ADD PRIMARY KEY (visitor_id, atelier_id);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE visitor_experts
    DROP CONSTRAINT visitor_experts_pkey;

ALTER TABLE visitor_experts
    ADD COLUMN id serial PRIMARY KEY;

-- +goose StatementEnd
