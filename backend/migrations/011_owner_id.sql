-- +goose Up
-- Владелец строки = users.id. Backfill v1: все бывшие общие строки → seed-admin.
-- UUID совпадает с internal/seed adminID. Таблиц org нет.

ALTER TABLE categories ADD COLUMN owner_id UUID REFERENCES users (id);
ALTER TABLE items ADD COLUMN owner_id UUID REFERENCES users (id);
ALTER TABLE audit_log ADD COLUMN owner_id UUID REFERENCES users (id);
ALTER TABLE notifications ADD COLUMN owner_id UUID REFERENCES users (id);

UPDATE categories SET owner_id = '11111111-1111-1111-1111-111111111111' WHERE owner_id IS NULL;
UPDATE items SET owner_id = '11111111-1111-1111-1111-111111111111' WHERE owner_id IS NULL;
UPDATE audit_log SET owner_id = '11111111-1111-1111-1111-111111111111' WHERE owner_id IS NULL;
UPDATE notifications SET owner_id = '11111111-1111-1111-1111-111111111111' WHERE owner_id IS NULL;

ALTER TABLE categories ALTER COLUMN owner_id SET NOT NULL;
ALTER TABLE items ALTER COLUMN owner_id SET NOT NULL;
ALTER TABLE audit_log ALTER COLUMN owner_id SET NOT NULL;
ALTER TABLE notifications ALTER COLUMN owner_id SET NOT NULL;

CREATE INDEX categories_owner_id_idx ON categories (owner_id);
CREATE INDEX items_owner_id_idx ON items (owner_id);
CREATE INDEX audit_log_owner_id_idx ON audit_log (owner_id);
CREATE INDEX notifications_owner_id_idx ON notifications (owner_id);

-- +goose Down
DROP INDEX IF EXISTS notifications_owner_id_idx;
DROP INDEX IF EXISTS audit_log_owner_id_idx;
DROP INDEX IF EXISTS items_owner_id_idx;
DROP INDEX IF EXISTS categories_owner_id_idx;

ALTER TABLE notifications DROP COLUMN owner_id;
ALTER TABLE audit_log DROP COLUMN owner_id;
ALTER TABLE items DROP COLUMN owner_id;
ALTER TABLE categories DROP COLUMN owner_id;
