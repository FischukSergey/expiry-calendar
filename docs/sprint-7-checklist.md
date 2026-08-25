# Sprint 7 Checklist

Источник: [`sprint-7-plan.md`](sprint-7-plan.md). Код — только после закрытия Sprint 6.

## 1) Модель и миграции

- [ ] Таблица `orgs` (id, name, created_at).
- [ ] Таблица `org_members` (org_id, user_id, role `admin`|`viewer`, UNIQUE(org_id, user_id)).
- [ ] Таблица `org_invites` (org_id, token_hash, expires_at, created_by, redeemed_at, redeemed_by).
- [ ] Колонка `org_id` на `categories`, `items`, `audit_log`, `notifications` (+ индексы).
- [ ] Backfill: одна демо-org, все строки v1 и seed-пользователи привязаны к ней.
- [ ] Новые строки без `org_id` невозможны (NOT NULL + FK).

## 2) Auth и членство

- [ ] `POST /auth/register` без инвайта: org + membership admin + пара токенов.
- [ ] Claims access: `sub`, `role` (из членства), `org_id`, `iss=duekeep`, `iat`, `exp`.
- [ ] `POST /auth/login` и `POST /auth/refresh` кладут тот же `org_id` (текущая org пользователя).
- [ ] `GET /me`: `id`, `email`, `role`, `org_id`, `org_name`.
- [ ] Права мутаций — по `org_members.role`, не по глобальному `users.role`.

## 3) Инвайты

- [ ] `POST /api/v1/org/invites` — только admin текущей org; ответ: token + expires_in (сырой token один раз).
- [ ] `POST /api/v1/org/invites/accept` — cookie/Bearer; viewer в org инвайта.
- [ ] `POST /auth/register` с `invite_token` — без новой org, сразу viewer.
- [ ] Просроченный / погашенный / неизвестный токен — 401 или 422 (как в контракте).
- [ ] Повторное членство той же org — 409, без второй строки.

## 4) Изоляция выборок

- [ ] List/get/patch/delete `items`, renew, bulk, CSV — только текущий `org_id`.
- [ ] CRUD `categories` — только текущий `org_id`; глубина ≤ 3 как в Sprint 2.
- [ ] `GET /audit` — только события своей org.
- [ ] Чужой UUID → `404 not_found` (не `403`).
- [ ] `item_kinds` остаются общими на инсталляцию (чтение всем auth).

## 5) Realtime и обзор

- [ ] `GET /dashboard`, `GET /calendar` — агрегаты только своей org.
- [ ] `GET /notifications` и read/read-all — только своя org.
- [ ] Тикер создаёт notification с `org_id` item.
- [ ] SSE: событие только клиентам с тем же `org_id` в токене.
- [ ] Web Push: не слать подписчику чужой org.

## 6) Seed и совместимость v1

- [ ] Демо-org: `admin@duekeep.local` (admin) + `viewer@duekeep.local` (viewer) + прежний seed items.
- [ ] Новая org при register: копия дефолтных категорий, без items демо.
- [ ] Повторный `compose up` не плодит org/members/invites.
- [ ] Login/refresh/logout контракта Sprint 2 не ломаются (поля только добавляются).

## 7) UI / PWA

- [ ] Профиль: название org и роль.
- [ ] Хозяин: создать инвайт, показать копируемую ссылку.
- [ ] Экран/шаг «принять инвайт» (токен в query или форма).
- [ ] Viewer по-прежнему без кнопок мутаций.
- [ ] Register без инвайта попадает в пустую свою область (не в демо-список).

## 8) Тесты

- [ ] Unit: хеш инвайта; register создаёт org; accept не создаёт вторую org.
- [ ] Isolation: два admin не видят чужие items / dashboard.
- [ ] Invite: viewer читает, `POST /items` → 403.
- [ ] Чужой id → 404.
- [ ] SSE/push не утекает в другой org (хотя бы на уровне service-фильтра).

## 9) DoD

- [ ] Контракт [`api-sprint-7.md`](api-sprint-7.md) соблюдён, `openapi.yaml` дополнен.
- [ ] [`known-limitations-sprint-7.md`](known-limitations-sprint-7.md) заполнен.
- [ ] Демо из плана §7 пройдено (curl или UI).
- [ ] `task lint` и `task test` зелёные.
