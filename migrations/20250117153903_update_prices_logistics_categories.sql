-- +goose Up
-- +goose StatementBegin
ALTER TABLE logistics_categories
ALTER COLUMN price TYPE REAL,
ALTER COLUMN price_member TYPE REAL;

UPDATE logistics_categories
SET price = 5, price_member = 2.5
WHERE cat_name = 'Halbe Palette';

UPDATE logistics_categories
SET price = 10, price_member = 5
WHERE cat_name = 'Ganze Palette';

UPDATE logistics_categories
SET price = 20, price_member = 10
WHERE cat_name = 'Doppelte Palette';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE logistics_categories
ALTER COLUMN price TYPE INT,
ALTER COLUMN price_member TYPE INT;

UPDATE logistics_categories
SET price = 10, price_member = 5
WHERE cat_name = 'Halbe Palette';

UPDATE logistics_categories
SET price = 15, price_member = 10
WHERE cat_name = 'Ganze Palette';

UPDATE logistics_categories
SET price = 30, price_member = 20
WHERE cat_name = 'Doppelte Palette';
-- +goose StatementEnd
