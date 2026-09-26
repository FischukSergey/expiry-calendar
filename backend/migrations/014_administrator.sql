-- +goose Up
ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'viewer', 'administrator'));

-- Справочник инсталляции. Повтор миграции и уже залитые slug не дублируются.
INSERT INTO item_kinds (id, slug, name, color, attr_schema) VALUES
    ('33333333-3333-3333-3333-333333333301', 'domain', 'Домен', '#3B82F6',
     '[{"key":"registrar","label":"Регистратор","type":"string","required":false},{"key":"auto_renew","label":"Автопродление","type":"boolean","required":false}]'),
    ('33333333-3333-3333-3333-333333333302', 'subscription', 'Подписки', '#8B5CF6',
     '[{"key":"seats","label":"Места","type":"number","required":false},{"key":"auto_renew","label":"Автопродление","type":"boolean","required":false}]'),
    ('33333333-3333-3333-3333-333333333303', 'rent', 'Аренда', '#F59E0B',
     '[{"key":"landlord","label":"Арендодатель","type":"string","required":false},{"key":"address","label":"Адрес","type":"string","required":false}]'),
    ('33333333-3333-3333-3333-333333333304', 'contract', 'Договор', '#6366F1',
     '[{"key":"counterparty","label":"Контрагент","type":"string","required":false},{"key":"contract_number","label":"Номер договора","type":"string","required":false}]'),
    ('33333333-3333-3333-3333-333333333305', 'insurance', 'Страховка', '#10B981',
     '[{"key":"policy_number","label":"Номер полиса","type":"string","required":false},{"key":"insurer","label":"Страховщик","type":"string","required":false}]'),
    ('33333333-3333-3333-3333-333333333306', 'license', 'Лицензия', '#EC4899',
     '[{"key":"license_key","label":"Ключ","type":"string","required":false},{"key":"seats","label":"Места","type":"number","required":false}]'),
    ('33333333-3333-3333-3333-333333333307', 'tax', 'Налог', '#EF4444',
     '[{"key":"tax_authority","label":"Инспекция","type":"string","required":false},{"key":"period","label":"Период","type":"string","required":false}]'),
    ('33333333-3333-3333-3333-333333333308', 'vehicle', 'Авто', '#14B8A6',
     '[{"key":"vin","label":"VIN","type":"string","required":false},{"key":"plate","label":"Госномер","type":"string","required":false}]'),
    ('33333333-3333-3333-3333-333333333310', 'mobile', 'Мобильная связь', '#0EA5E9',
     '[{"key":"phone","label":"Номер телефона","type":"string","required":false},{"key":"operator","label":"Оператор","type":"string","required":false}]'),
    ('33333333-3333-3333-3333-333333333309', 'other', 'Прочее', '#6B7280', '[]')
ON CONFLICT (slug) DO NOTHING;

-- +goose Down
UPDATE users SET role = 'admin' WHERE role = 'administrator';
ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'viewer'));
