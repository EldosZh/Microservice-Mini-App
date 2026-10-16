CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    price_cents BIGINT NOT NULL CHECK (price_cents > 0),
    stock INTEGER NOT NULL CHECK (stock >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
