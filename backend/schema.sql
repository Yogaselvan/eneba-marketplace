CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS users (
  id            BIGSERIAL PRIMARY KEY,
  email         TEXT NOT NULL UNIQUE,
  name          TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS products (
  id               BIGSERIAL PRIMARY KEY,
  slug             TEXT NOT NULL UNIQUE,
  name             TEXT NOT NULL,
  platform         TEXT NOT NULL,
  region           TEXT NOT NULL,
  genre            TEXT NOT NULL,
  description      TEXT NOT NULL DEFAULT '',
  list_price_cents INT  NOT NULL,
  release_year     INT  NOT NULL DEFAULT 2024,
  sales            INT  NOT NULL DEFAULT 0,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS products_name_trgm ON products USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS products_platform ON products (platform);

CREATE TABLE IF NOT EXISTS sellers (
  id           BIGSERIAL PRIMARY KEY,
  name         TEXT NOT NULL UNIQUE,
  rating       NUMERIC(3,2) NOT NULL DEFAULT 4.5,
  sales_count  INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS offers (
  id          BIGSERIAL PRIMARY KEY,
  product_id  BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  seller_id   BIGINT NOT NULL REFERENCES sellers(id),
  price_cents INT NOT NULL,
  stock       INT NOT NULL DEFAULT 0 CHECK (stock >= 0)
);
CREATE INDEX IF NOT EXISTS offers_product ON offers (product_id);

CREATE TABLE IF NOT EXISTS orders (
  id          BIGSERIAL PRIMARY KEY,
  user_id     BIGINT NOT NULL REFERENCES users(id),
  total_cents INT NOT NULL,
  status      TEXT NOT NULL DEFAULT 'paid',
  payment_ref TEXT NOT NULL DEFAULT '',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS order_items (
  id          BIGSERIAL PRIMARY KEY,
  order_id    BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  offer_id    BIGINT NOT NULL REFERENCES offers(id),
  product_id  BIGINT NOT NULL REFERENCES products(id),
  price_cents INT NOT NULL
);

-- One row per sellable digital key. order_item_id IS NULL means still available.
CREATE TABLE IF NOT EXISTS keys (
  id            BIGSERIAL PRIMARY KEY,
  offer_id      BIGINT NOT NULL REFERENCES offers(id) ON DELETE CASCADE,
  code          TEXT NOT NULL,
  order_item_id BIGINT REFERENCES order_items(id)
);
CREATE INDEX IF NOT EXISTS keys_available ON keys (offer_id) WHERE order_item_id IS NULL;

CREATE TABLE IF NOT EXISTS banners (
  id        BIGSERIAL PRIMARY KEY,
  title     TEXT NOT NULL,
  subtitle  TEXT NOT NULL,
  cta       TEXT NOT NULL DEFAULT 'Shop now',
  link      TEXT NOT NULL,
  color_a   TEXT NOT NULL,
  color_b   TEXT NOT NULL,
  position  INT NOT NULL DEFAULT 0
);
