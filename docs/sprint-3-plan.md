# Sprint 3 — записи

Источник: [`ARCHITECTURE.md`](../ARCHITECTURE.md) §6, §8; [`api-sprint-2.md`](api-sprint-2.md).

## 1) Цель

CRUD записей обязательств: колонки + `attrs` JSONB, фильтры, пагинация, продление, bulk, журнал аудита.

## 2) Входные условия

Sprint 2: JWT, kinds, categories, seed пользователей и справочников.

## 3) Границы

### Входит

- таблица `items`, `renewals`, `audit_log`;
- валидация колонок и `attrs` по `attr_schema`;
- список: q, kind, status, category (+ потомки), vendor, даты, цена, billing, tag, sort;
- `POST /items/{id}/renew`;
- `POST /items/bulk`;
- `GET /audit` (admin);
- запись audit на create/update/delete/renew/bulk.

### Не входит

- тикер статусов и автопереход expiring/expired (ручной status или упрощённый расчёт **при записи** — да);
- SSE, push, dashboard, calendar, CSV, UI.

**Статус при записи:** service считает `active`/`expiring`/`expired` от `Clock.Today()` если статус не `cancelled`/`archived`. Тикер (повторный пересчёт без PATCH) — Sprint 4.

## 4) Backlog

### A. Миграции items / renewals / audit_log, индексы из архитектуры.

### B. Items service

- инварианты дат и cost;
- attrs: лишние ключи → 422;
- запрет viewer на мутации.

### C. List + pagination `page`/`per_page`/`total`.

### D. Renew и bulk + audit.

### E. Тесты: CRUD, 403 viewer, attrs, фильтр+пагинация, renew пишет историю.

## 5) Техрешения

- Гибрид: фильтры только по колонкам, не по JSONB.
- `category_id` в фильтре включает потомков (один запрос или CTE).
- Audit `before_json` / `after_json` краткие, без паролей (их и нет в item).

## 6) DoD

- admin создаёт/правит/удаляет/продлевает;
- список фильтруется и пагинируется;
- viewer только читает;
- audit виден admin;
- [`api-sprint-3.md`](api-sprint-3.md).

## 7) Демо

Создать аренду и подписку, продлить одну, отфильтровать `kind=rent`, открыть audit.

## 8) Риски

- N+1 на дерево категорий в фильтре → closure/CTE сразу.
- Пустой service-прокси → вся валидация attrs/status в service.

## 9) Артефакты

- API записей, тесты, docs.
