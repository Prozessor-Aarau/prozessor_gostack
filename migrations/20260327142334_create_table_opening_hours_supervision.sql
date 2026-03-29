-- +goose Up
-- +goose StatementBegin
CREATE TABLE opening_hours (
    id serial PRIMARY KEY,
    start_date timestamptz NOT NULL,
    end_date timestamptz NOT NULL,
    created_on timestamptz DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE opening_hours_supervision (
    opening_hour_id int REFERENCES opening_hours (id) ON DELETE CASCADE,
    visitor_id uuid REFERENCES visitors (id) ON DELETE CASCADE,
    PRIMARY KEY (opening_hour_id, visitor_id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS opening_hours_supervision;

DROP TABLE IF EXISTS opening_hours;

-- +goose StatementEnd
