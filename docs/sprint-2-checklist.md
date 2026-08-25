# Sprint 2 Checklist

Источник: [`sprint-2-plan.md`](sprint-2-plan.md).

## 1) Данные

- [x] Миграция `users`.
  Примечание: `002_users.sql` — citext, `email` UNIQUE, `role` admin/viewer.
- [x] Миграция `refresh_tokens` (family_id, token_hash, revoked_at).
  Примечание: `003_refresh_tokens.sql` + индексы user_id / family_id / expires_at.
- [x] Миграция `item_kinds`.
  Примечание: `004_item_kinds.sql` — slug UNIQUE, `attr_schema` JSONB `[]`.
- [x] Миграция `categories`.
  Примечание: `005_categories.sql` — self-FK `parent_id`.
- [x] Идемпотентный seed: 2 пользователя, 9 kinds, ≥ 10 категорий.
  Примечание: `seed.Run` после goose. Повторный restart: 2 / 9 / 13 строк, без дублей. Пароли в README.

## 2) Auth

- [ ] `POST /api/v1/auth/register`
- [ ] `POST /api/v1/auth/login`
- [ ] `POST /api/v1/auth/refresh` (cookie или body)
- [ ] `POST /api/v1/auth/logout`
- [ ] `POST /api/v1/auth/logout-all`
- [ ] `GET /api/v1/me`
- [ ] Middleware Bearer, 401 без токена
- [ ] Ротация refresh
- [ ] Reuse → revoke family

## 3) Справочники

- [ ] `GET/POST /api/v1/kinds`, `PATCH/DELETE /api/v1/kinds/{id}`
- [ ] `attr_schema` валидируется как массив описателей
- [ ] `GET/POST /api/v1/categories`, `PATCH/DELETE /api/v1/categories/{id}`
- [ ] Глубина ≤ 3
- [ ] Запрет удалить категорию с детьми
- [ ] Viewer: только чтение

## 4) Тесты

- [ ] Unit: hash refresh, claims, глубина дерева
- [ ] Integration: login → refresh → logout; viewer 403 на запись kind

## 5) DoD

- [ ] Контракт [`api-sprint-2.md`](api-sprint-2.md) соблюдён
- [ ] [`known-limitations-sprint-2.md`](known-limitations-sprint-2.md) заполнен
