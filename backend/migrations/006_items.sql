-- +goose Up
CREATE TABLE items (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title              TEXT NOT NULL,
    description        TEXT NOT NULL DEFAULT '',
    kind_id            UUID NOT NULL REFERENCES item_kinds(id),
    category_id        UUID NULL REFERENCES categories(id),
    vendor             TEXT NOT NULL DEFAULT '',
    tags               TEXT[] NOT NULL DEFAULT '{}',
    cost_amount        INT NOT NULL DEFAULT 0 CHECK (cost_amount >= 0),
    currency           CHAR(3) NOT NULL,
    billing_period     TEXT NOT NULL CHECK (billing_period IN ('one_time', 'monthly', 'yearly')),
    started_at         DATE NULL,
    expires_at         DATE NOT NULL,
    notify_before_days INT NOT NULL DEFAULT 30 CHECK (notify_before_days >= 0),
    url                TEXT NOT NULL DEFAULT '',
    account_hint       TEXT NOT NULL DEFAULT '',
    status             TEXT NOT NULL CHECK (status IN ('active', 'expiring', 'expired', 'cancelled', 'archived')),
    attrs              JSONB NOT NULL DEFAULT '{}',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (started_at IS NULL OR started_at <= expires_at)
);

CREATE INDEX items_expires_at_idx ON items (expires_at);
CREATE INDEX items_status_idx ON items (status);
CREATE INDEX items_kind_id_idx ON items (kind_id);
CREATE INDEX items_category_id_idx ON items (category_id);
CREATE INDEX items_currency_idx ON items (currency);
CREATE INDEX items_tags_gin ON items USING GIN (tags);
CREATE INDEX items_attrs_gin ON items USING GIN (attrs);

-- +goose Down
DROP TABLE IF EXISTS items;
