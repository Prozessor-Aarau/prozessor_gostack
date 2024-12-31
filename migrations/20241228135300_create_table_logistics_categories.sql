-- +goose Up
-- +goose StatementBegin
CREATE TABLE logistics_categories (
  id SERIAL PRIMARY KEY,
  cat_name TEXT NOT NULL,
  price INT NOT NULL,
  price_member INT NOT NULL,
  updated_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
  created_on TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

INSERT INTO logistics_categories (cat_name, price, price_member)
VALUES ('Halbe Palette', 10, 5),
      ('Ganze Palette', 15, 10),
      ('Doppelte Palette', 30, 20),
      ('Sperrgut', 0, 0),
      ('Kleingut', 0, 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE logistics_categories;
-- +goose StatementEnd
