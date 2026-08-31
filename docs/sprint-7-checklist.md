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
  Примечание: чужой id → 404. Списки ещё общие (п.3).

## 3) Изоляция выборок

- [ ] List/get/patch/delete `items`, renew, bulk, CSV — только свой `owner_id`.
- [ ] CRUD `categories` — только свой `owner_id`; глубина ≤ 3 как в Sprint 2.
- [ ] `GET /audit` — только свои события.
- [ ] Чужой UUID → `404 not_found` (не `403`).
- [ ] `item_kinds` остаются общими на инсталляцию (чтение всем auth).

## 4) Realtime и обзор

- [ ] `GET /dashboard`, `GET /calendar` — агрегаты только своих items.
- [ ] `GET /notifications` и read/read-all — только свои.
- [ ] Тикер создаёт notification с `owner_id` владельца item.
- [ ] SSE: событие только клиентам с тем же `sub`.
- [ ] Web Push: не слать подписчику чужие items.

## 5) Seed и прод

- [ ] Prod compose: seed не выполняется.
- [ ] Local seed: каталог 50+ на seed-admin; общего каталога и шаринга с viewer нет.
- [ ] Register: копия дефолтных категорий, без items seed.
- [ ] Повторный `compose up` не плодит seed-пользователей и категории.
- [ ] Login/refresh/logout контракта Sprint 2 не ломаются (кроме роли после register: теперь `admin`).

## 6) UI / PWA

- [ ] Register попадает в пустой свой список (не в seed).
- [ ] Нет экранов org / инвайта.
- [ ] Профиль без названия org.

## 7) Тесты

- [ ] Isolation: два admin не видят чужие items / dashboard.
- [ ] Register не видит seed-каталог.
- [ ] Чужой id → 404.
- [ ] SSE/push не утекает другому пользователю (хотя бы на уровне service-фильтра).

## 8) DoD

- [ ] Контракт [`api-sprint-7.md`](api-sprint-7.md) соблюдён, `openapi.yaml` дополнен.
- [ ] [`known-limitations-sprint-7.md`](known-limitations-sprint-7.md) заполнен.
- [ ] Демо из плана §7 пройдено (curl или UI).
- [ ] `task lint` и `task test` зелёные.
