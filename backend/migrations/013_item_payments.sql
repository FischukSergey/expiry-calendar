-- +goose Up
-- Факт оплаты вхождения: строка есть = оплачено. Не полная серия дат.
CREATE TABLE item_payments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id    UUID NOT NULL REFERENCES items (id) ON DELETE CASCADE,
    owner_id   UUID NOT NULL REFERENCES users (id),
    paid_on    DATE NOT NULL,
    amount     INT NOT NULL CHECK (amount >= 0),
    currency   CHAR(3) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (item_id, paid_on)
);

CREATE INDEX item_payments_owner_paid_on_idx ON item_payments (owner_id, paid_on);

-- Снимок суммы на якоре. items.status = paid не затираем (заморозка записи).
INSERT INTO item_payments (item_id, owner_id, paid_on, amount, currency)
SELECT id, owner_id, expires_at, cost_amount, currency
FROM items
WHERE status = 'paid'
ON CONFLICT (item_id, paid_on) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS item_payments;
