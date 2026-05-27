CREATE FUNCTION set_updated_at_column_to_now()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at=now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS organisations (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           BIGINT NOT NULL,
    name              TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);

CREATE TRIGGER set_organisations_updated_at BEFORE UPDATE ON organisations FOR EACH ROW EXECUTE PROCEDURE set_updated_at_column_to_now();

CREATE TABLE IF NOT EXISTS purchases (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           BIGINT NOT NULL,
    purchase_date     DATE NOT NULL,
    organisation_name TEXT NOT NULL,
    description       TEXT,
    source_type       TEXT NOT NULL CHECK (source_type IN ('manual', 'text', 'photo', 'qr')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER set_purchases_updated_at BEFORE UPDATE ON purchases FOR EACH ROW EXECUTE PROCEDURE set_updated_at_column_to_now();

CREATE TABLE IF NOT EXISTS categories (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              TEXT NOT NULL UNIQUE,
    icon              TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS expenses (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    purchase_id       UUID NOT NULL REFERENCES purchases(id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    quantity          NUMERIC(12,3) NOT NULL,
    unit_price_minor  BIGINT NOT NULL,
    total_price_minor BIGINT NOT NULL,
    currency          TEXT NOT NULL,
    category_id       UUID NOT NULL REFERENCES categories(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER set_expenses_updated_at BEFORE UPDATE ON expenses FOR EACH ROW EXECUTE PROCEDURE set_updated_at_column_to_now();

CREATE INDEX IF NOT EXISTS idx_purchases_user_date ON purchases(user_id, purchase_date DESC);
CREATE INDEX IF NOT EXISTS idx_organisations_user_updated_at ON organisations(user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_expenses_purchase_id ON expenses(purchase_id);
