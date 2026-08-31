# Sprint 7 Checklist

Источник: [`sprint-7-plan.md`](sprint-7-plan.md). Код — только после закрытия Sprint 6.

## 1) Модель и миграции

- [x] Колонка `owner_id` на `categories`, `items`, `audit_log`, `notifications` (FK на `users`, индексы).
  Примечание: `011_owner_id.sql`. В API поле не отдаём (`json:"-"`).
- [x] Backfill: бывшие общие строки seed → `owner_id` seed-admin.
  Примечание: UPDATE на UUID `11111111-…`; новые seed-INSERT тоже пишут его.
- [x] Новые строки без `owner_id` невозможны (NOT NULL + FK).
- [x] Нет таблиц `orgs` / `org_members` / `org_invites` и нет колонки `org_id`.

## 2) Auth

- [x] `POST /auth/register` создаёт `admin` (не `viewer`) и пару токенов.
- [x] Claims access без `org_id`: `sub`, `role`, `iss=duekeep`, `iat`, `exp`.
- [x] `POST /auth/login` и `POST /auth/refresh` — тот же набор claims.
- [x] `GET /me`: `id`, `email`, `role` (без `org_id` / `org_name`).
- [x] Мутации предметных сущностей — только если `owner_id` = текущий `sub`.
  Примечание: чужой id → 404.

## 3) Изоляция выборок

- [x] List/get/patch/delete `items`, renew, bulk, CSV — только свой `owner_id`.
  Примечание: `ItemFilter.OwnerID` из sub; get/list/export/audit scoped. Dashboard — п.4.
- [x] CRUD `categories` — только свой `owner_id`; глубина ≤ 3 как в Sprint 2.
- [x] `GET /audit` — только свои события.
- [x] Чужой UUID → `404 not_found` (не `403`).
- [x] `item_kinds` остаются общими на инсталляцию (чтение всем auth).

## 4) Realtime и обзор

- [x] `GET /dashboard`, `GET /calendar` — агрегаты только своих items.
- [x] `GET /notifications` и read/read-all — только свои.
- [x] Тикер создаёт notification с `owner_id` владельца item.
  Примечание: `TICKER_EVERY` по умолчанию 12h (статус — день UTC); Tick сразу при старте.
- [x] SSE: событие только клиентам с тем же `sub`.
- [x] Web Push: не слать подписчику чужие items.

## 5) Seed и прод

- [x] Prod compose: seed не выполняется.
  Примечание: `SEED=false` в `deploy/prod`; `seed.Run` не вызывается. Справочник kinds — `EnsureKinds`.
- [x] Local seed: каталог 50+ на seed-admin; общего каталога и шаринга с viewer нет.
  Примечание: items `owner_id` = seed-admin; viewer — локальный стенд без шаринга.
- [x] Register: копия дефолтных категорий, без items seed.
- [x] Повторный `compose up` не плодит seed-пользователей и категории.
  Примечание: `ON CONFLICT` по email/id/slug; повторный INSERT не плодит строки.
- [x] Login/refresh/logout контракта Sprint 2 не ломаются (кроме роли после register: теперь `admin`).
  Примечание: `TestLoginRefreshLogout`, `TestRefreshRotationAndReuse`.

## 6) UI / PWA

- [x] Register попадает в пустой свой список (не в seed).
  Примечание: после register `GuestRoute` ведёт на `/items`; пустой стейт без seed-копирайта.
- [x] Нет экранов org / инвайта.
  Примечание: маршрутов `/org` и инвайта нет; с экрана входа убрана «роль viewer».
- [x] Профиль без названия org.
  Примечание: только email; viewer — пометка локального стенда, не продуктовая роль.

## 7) Тесты

- [x] Isolation: два admin не видят чужие items / dashboard.
- [x] Register не видит seed-каталог.
  Примечание: `TestRegisterDoesNotSeeSeedCatalog` — пустой list/dashboard, GET seed и item A → 404 у B.
- [x] Чужой id → 404.
- [x] SSE/push не утекает другому пользователю (хотя бы на уровне service-фильтра).

## 8) DoD

- [x] Контракт [`api-sprint-7.md`](api-sprint-7.md) соблюдён, `openapi.yaml` дополнен.
  Примечание: register → admin + категории; /me без org; нет `/org`; `owner_id` не в JSON.
- [x] [`known-limitations-sprint-7.md`](known-limitations-sprint-7.md) заполнен.
- [x] Демо из плана §7 пройдено (curl или UI).
  Примечание: шаги 1–2 — `TestRegisterDoesNotSeeSeedCatalog`. Шаг 3 (прод без seed) — `SEED=false` + `TestSeedEnabled`.
- [x] `task lint` и `task test` зелёные.
