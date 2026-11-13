-- This file is embedded into the Go binary (see db/postgres.go) and run on startup.
-- "IF NOT EXISTS" makes it safe to run many times.

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    name          TEXT        NOT NULL,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'customer', -- 'customer' or 'admin'
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Money is stored as an integer number of cents (never use float for money!).
-- The CHECK constraints are a last line of defence: the DB refuses negative stock.
CREATE TABLE IF NOT EXISTS products (
    id          SERIAL PRIMARY KEY,
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    price_cents BIGINT      NOT NULL CHECK (price_cents >= 0),
    stock       INT         NOT NULL DEFAULT 0 CHECK (stock >= 0),
    category    TEXT        NOT NULL DEFAULT '',
    image_url   TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cart_items (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity   INT NOT NULL CHECK (quantity > 0),
    UNIQUE (user_id, product_id) -- one row per product per user
);

CREATE TABLE IF NOT EXISTS orders (
    id          SERIAL PRIMARY KEY,
    user_id     INT         NOT NULL REFERENCES users(id),
    total_cents BIGINT      NOT NULL,
    status      TEXT        NOT NULL DEFAULT 'pending', -- pending, paid, shipped, delivered, cancelled
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- We copy name and price into the order item, so the order history
-- stays correct even if the admin later edits or deletes the product.
CREATE TABLE IF NOT EXISTS order_items (
    id           SERIAL PRIMARY KEY,
    order_id     INT    NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id   INT    REFERENCES products(id) ON DELETE SET NULL,
    product_name TEXT   NOT NULL,
    price_cents  BIGINT NOT NULL,
    quantity     INT    NOT NULL CHECK (quantity > 0)
);

CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_orders_user_id    ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);
