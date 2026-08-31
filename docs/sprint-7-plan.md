# Sprint 7 — свои данные (прод)

Источник: решение после v1; спринты 1–6 не переписываем. Сдача защиты — тег `v1.0.0`.

## 1) Цель

Прод (`duekeep.ru`): **каждый регистрируется и видит только своё**. Один пользователь — хозяин своих записей. Роль `viewer`, шаринг и изоляция по `org_id` **не делаем**.

К концу спринта:

- `POST /auth/register` создаёт `admin` с пустым каталогом (копия дефолтных категорий, без seed-items);
- предметные выборки и мутации — только строки владельца (`owner_id` = `sub` из access);
- JWT без `org_id`: claims как в Sprint 2 (`sub`, `role`, `iss`, `iat`, `exp`);
- на проде seed выключен; локально seed-items принадлежат seed-admin, общий каталог v1 больше не используется;
- два независимых аккаунта не видят записи друг друга.

## 2) Входные условия

Sprint 6 закрыт, тег `v1.0.0`. v1: общий набор данных, роли admin/viewer на инсталляцию.

## 3) Границы

### Входит

- колонка `owner_id` (FK на `users`) на `items`, `categories`, `audit_log`, `notifications`;
- register → роль `admin`, свои категории, без чужих items;
- все list/get/patch/delete/renew/bulk/CSV/dashboard/calendar/notifications scoped по `owner_id`;
- SSE, тикер, Web Push — только события своих записей;
- флаг/режим без seed на prod compose;
- UI: register сразу в пустой свой список; кнопки мутаций у admin (единственная роль после register).

### Не входит

- таблицы `orgs`, `org_members`, `org_invites` и поле `org_id`;
- инвайт-ссылки, шаринг, роль `viewer` в новой модели;
- несколько пространств / переключение org;
- почта, Telegram, биллинг, квоты;
- per-user справочник `item_kinds` (остаётся общим на инсталляцию);
- второй инстанс backend;
- переписывание контрактов спринтов 1–6 «задним числом»: меняем поведение register и добавляем scope; пути ручек те же.

## 4) Backlog

### A. Схема

- `owner_id NOT NULL` + FK + индексы на `categories`, `items`, `audit_log`, `notifications` (renewals — через item).
- Backfill локального seed: все бывшие общие строки → `owner_id` seed-admin.
- Новая строка без `owner_id` невозможна.

### B. Auth

- `register` больше не создаёт `viewer`: новый пользователь — `admin`.
- login/refresh без новых claims.
- `GET /me` — как Sprint 2 (`id`, `email`, `role`), без `org_id`.
- Права мутаций: auth + владение строкой. Глобальный `users.role` для новых аккаунтов — `admin`.

### C. Изоляция API

- repository: обязательный `owner_id` во всех запросах предметных таблиц.
- Чужой UUID → `404 not_found` (не `403`).
- `item_kinds` общие; кто пишет kinds — зафиксировать в limitations (не каждый зарегистрированный).

### D. Realtime и обзор

- тикер, SSE, Web Push, dashboard, calendar, CSV — только items текущего `sub`.

### E. Seed и прод

- Prod: seed не запускать (`SEED` / аналог в prod compose).
- Local: seed-admin + каталог 50+ на нём; seed-viewer не обязателен (шаринга нет).
- Register: копия дефолтного дерева категорий, kinds не копировать.
- Повторный `compose up` не плодит пользователей и категории seed.

### F. UI

- Register → пустой свой список.
- Профиль без org и без инвайтов.
- Скрыть или убрать сценарий «viewer без кнопок» как продуктовую роль (остаётся только если локальный seed-viewer ещё жив).

### G. Тесты

- два register не видят чужие items / dashboard;
- чужой id → 404;
- register не видит seed-каталог;
- SSE/push не утекает другому `sub`.

## 5) Техрешения

- Один сервер, один Postgres: изоляция = `WHERE owner_id = $sub`. Не `org_id`.
- Клиент — PWA; офлайн-CRUD по-прежнему нет.
- Интерфейсы repo по-прежнему в service.
- Неизвестный refresh — 401 без revoke family.

## 6) DoD

- два независимых аккаунта не видят записи друг друга;
- register даёт admin и пустой свой каталог;
- login/refresh как в Sprint 2, без `org_id` в claims;
- на prod compose seed выключен;
- контракт [`api-sprint-7.md`](api-sprint-7.md).

## 7) Демо

1. Register `a@…` — пустой список, свои категории.
2. Создать item. Register `b@…` — item A не виден, `GET` по id A → 404.
3. Прод: чистый том, без `admin@duekeep.local` и без 50+ seed.

## 8) Риски

- забыть `owner_id` в одном запросе → утечка между пользователями;
- SSE/тикер шлют всем сокетам, как в Sprint 4 («данные общие»);
- kinds общие: не дать каждому admin ломать глобальный справочник;
- register в v1 был viewer — клиент и тесты Sprint 2, которые ждут viewer, поправить только аддитивно в этом спринте.

## 9) Артефакты

- миграции `owner_id` + backfill seed;
- scoped queries;
- seed off на prod;
- тесты изоляции по пользователю;
- docs спринта.
