---
name: duekeep
description: >-
  Duekeep (expiry-calendar) domain invariants, layers, and Sprint 7 isolation.
  Use when analyzing or changing backend/frontend of this repo, implementing
  sprint checklist items, touching owner_id, auth, items, categories, ticker,
  SSE, push, dashboard, or seed.
---

# Duekeep — как устроен проект сейчас

Читай чеклист и код, не устаревшие абзацы ARCHITECTURE про «общий каталог v1» как текущую модель API.

## Статус

- Спринты **1–6 закрыты** (тег сдачи `v1.0.0`: общий каталог, admin/viewer).
- Текущий фокус — **Sprint 7**: свои данные. §1–6 закрыты. Дальше тесты/DoD.
- Sprint 8 (CD) — после 7. Не начинай его сам.

Документы спринта N: `docs/sprint-N-plan.md`, `checklist`, `api-sprint-N.md`, `known-limitations-sprint-N.md`. Журнал — `REPORT.md`.

## Стек

Go 1.25, модуль `duekeep`, один бинарь `cmd/server`. Слои: `handler → service → repository`. Интерфейсы repo — в `service`.

chi, pgx/v5, goose (SQL embed, `001`…`011_owner_id.sql`), slog JSON, OpenAPI `backend/openapi.yaml` + `/docs`. Frontend: Vite 6, React 18, TS, Tailwind 4, PWA, nginx `:80` → API `:8080`. Postgres с хоста `localhost:15432`. Compose-проект `duekeep`.

Качество: только `task lint` / `task test`. `docker compose` с пробелом.

## Auth

- JWT access: `sub`, `role`, `iss=duekeep`, `iat`, `exp`. **Нет `org_id`.**
- Refresh: JSON body и/или cookie `duekeep_refresh` (body важнее). Ротация, `family_id`, reuse → revoke family.
- `POST /auth/register` → роль **`admin`**, пара токенов. Viewer остаётся у локального seed, не у новых аккаунтов.
- `GET /me`: `id`, `email`, `role`.

## Изоляция (`owner_id`)

Таблицы с `owner_id NOT NULL` + FK на `users`: `categories`, `items`, `audit_log`, `notifications`. В API не отдаём.

Правило: выборка и мутация предметных строк — `owner_id = sub`. Чужой UUID → **404**, не 403. Scope ставит service из `middleware.UserID`, не из query/body.

| Поверхность | Как |
|---|---|
| items list/export | `ItemFilter.OwnerID` |
| items get/patch/delete/renew/bulk | `requireOwner` после `ByID` |
| categories | `List(ownerID)`; create/patch parent только в своём дереве; глубина ≤ 3 |
| audit, notifications | `List`/`MarkRead`/`MarkAllRead` с owner |
| dashboard, calendar | `ListOpenByOwner` |
| SSE | `Hub.Subscribe(userID)`, `Notify` только тому же `sub` |
| push | `Broadcast` если `sub.UserID == n.OwnerID` |

**Не фильтровать** `Items.ListOpen` — тикер должен видеть все open items всех владельцев.

**Общее на инсталляцию:** `item_kinds`. GET всем auth. Не делать per-user kinds.

Create/import/ticker notification: писать `OwnerID`. Пустой owner в INSERT notification ломает uuid NOT NULL.

## Тикер

`TICKER_EVERY` (дефолт `12h`) + Tick при старте. Статус по `Clock.Today` (день UTC). Не возвращать интервал 60 с без явной просьбы.

## Ещё не сделано (Sprint 7 §7–8)

- Тест: register не видит seed-каталог. DoD §8.

## Не делать

`org_id`, таблицы org/invites, шаринг, почта, Telegram, вложения, iCal, офлайн-CRUD, конвертация валют, фильтр по JSONB attrs, второй инстанс backend, GORM/Redis/Kafka/WebSocket.

## Документы при смене API

Вместе с хендлерами: `docs/api-sprint-7.md` и `backend/openapi.yaml`. Чеклист `[x]` только после проверки. DoD §8 — после lint/test/демо.
