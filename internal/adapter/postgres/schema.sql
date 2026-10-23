CREATE TABLE IF NOT EXISTS users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'disabled')),
    token_version INT NOT NULL DEFAULT 0,
    failed_logins INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS verification_codes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    purpose TEXT NOT NULL CHECK (purpose IN ('confirm', 'reset')),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    attempts INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS products (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    business_id TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL,
    price_ht BIGINT NOT NULL CHECK (price_ht >= 0),
    vat_rate BIGINT NOT NULL DEFAULT 2000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS baskets (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    business_id TEXT NOT NULL UNIQUE,
    user_id     BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS basket_items (
    basket_id  BIGINT NOT NULL REFERENCES baskets(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity   INT NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (basket_id, product_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    business_id   TEXT NOT NULL UNIQUE,
    user_id       BIGINT NOT NULL REFERENCES users(id),
    status        TEXT NOT NULL CHECK (status IN ('pending', 'paid', 'shipping', 'delivered', 'cancelled')),
    cancel_reason TEXT NOT NULL DEFAULT '',
    total_ttc     BIGINT NOT NULL CHECK (total_ttc >= 0),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS orders_user_id_idx ON orders (user_id);

CREATE TABLE IF NOT EXISTS order_items (
    order_id            BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id          BIGINT NOT NULL REFERENCES products(id),
    product_business_id TEXT NOT NULL,
    name                TEXT NOT NULL,
    unit_price_ttc      BIGINT NOT NULL CHECK (unit_price_ttc >= 0),
    quantity            INT NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (order_id, product_id)
);

CREATE TABLE IF NOT EXISTS payments (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id   BIGINT NOT NULL UNIQUE REFERENCES orders(id) ON DELETE CASCADE,
    amount     BIGINT NOT NULL CHECK (amount >= 0),
    card_last4 TEXT NOT NULL CHECK (card_last4 ~ '^[0-9]{4}$'), -- vérifie que le champ contient exactement 4 chiffres
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);