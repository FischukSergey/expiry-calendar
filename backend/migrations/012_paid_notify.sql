-- +goose Up
ALTER TABLE items DROP CONSTRAINT items_status_check;
ALTER TABLE items ADD CONSTRAINT items_status_check
    CHECK (status IN ('active', 'expiring', 'expired', 'cancelled', 'archived', 'paid'));
ALTER TABLE items ALTER COLUMN notify_before_days DROP NOT NULL;

-- +goose Down
UPDATE items SET notify_before_days = 30 WHERE notify_before_days IS NULL;
UPDATE items SET status = 'active' WHERE status = 'paid';
ALTER TABLE items ALTER COLUMN notify_before_days SET NOT NULL;
ALTER TABLE items DROP CONSTRAINT items_status_check;
ALTER TABLE items ADD CONSTRAINT items_status_check
    CHECK (status IN ('active', 'expiring', 'expired', 'cancelled', 'archived'));
