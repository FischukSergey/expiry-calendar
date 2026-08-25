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

- [x] `POST /api/v1/auth/register`
  Примечание: всегда viewer; email UNIQUE → 409; пароль короче 8 символов → 422.
- [x] `POST /api/v1/auth/login`
  Примечание: неверный email/пароль → 401 без различия причины. Cookie `duekeep_refresh`.
- [x] `POST /api/v1/auth/refresh` (cookie или body)
  Примечание: непустой body важнее cookie.
- [x] `POST /api/v1/auth/logout`
  Примечание: access и/или refresh; неизвестный refresh → 204.
- [x] `POST /api/v1/auth/logout-all`
  Примечание: только access; гасит все семьи.
- [x] `GET /api/v1/me`
  Примечание: `{id,email,role}`.
- [x] Middleware Bearer, 401 без токена
  Примечание: на `/me` и `/logout-all`. Register/login/refresh без Bearer.
- [x] Ротация refresh
  Примечание: тот же `family_id`, старый хеш revoked.
- [x] Reuse → revoke family
  Примечание: отозванный/прокрученный refresh гасит семью; неизвестный токен — 401 без revoke.

## 3) Справочники

- [x] `GET/POST /api/v1/kinds`, `PATCH/DELETE /api/v1/kinds/{id}`
  Примечание: GET под Bearer; мутации — admin. DELETE занятого kind → 409 (count items = 0 до Sprint 3).
- [x] `attr_schema` валидируется как массив описателей
  Примечание: `{key,label,type,required}`, type ∈ string|number|boolean, key уникален.
- [x] `GET/POST /api/v1/categories`, `PATCH/DELETE /api/v1/categories/{id}`
  Примечание: GET отдаёт `{items:[корни]}` с вложенными `children`.
- [x] Глубина ≤ 3
  Примечание: корень = 1; create/move глубже → 422. Цикл при смене родителя → 422.
- [x] Запрет удалить категорию с детьми
  Примечание: дети или items → 409.
- [x] Viewer: только чтение
  Примечание: `RequireAdmin` на POST/PATCH/DELETE; viewer → 403, не 401.

## 4) Тесты

- [x] Unit: hash refresh, claims, глубина дерева
  Примечание: `TestHashRefreshStable` (SHA-256 hex), `TestParseAccessClaims` (sub/role/iss/iat/exp), `CategoryDepth` 1/3/−1/цикл.
- [x] Integration: login → refresh → logout; viewer 403 на запись kind
  Примечание: httptest + in-memory store (`flow_test.go`). Register→refresh→logout 401; reuse гасит family; viewer GET kinds/categories 200, POST kind 403; admin login + POST kind 201.

## 5) DoD

- [x] Контракт [`api-sprint-2.md`](api-sprint-2.md) соблюдён
  Примечание: auth/kinds/categories совпадают с хендлерами и `openapi.yaml`. Сценарии плана §6 покрыты `flow_test` + service-тестами.
- [x] [`known-limitations-sprint-2.md`](known-limitations-sprint-2.md) заполнен
  Примечание: access после logout, CountItems=0, нет UI логина, integration без Postgres.
