CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    total_cents BIGINT NOT NULL CHECK (total_cents >= 0),
    status TEXT NOT NULL CHECK (status IN ('created', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
