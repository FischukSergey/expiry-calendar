# API контракты Sprint 4

Предыдущее: [`api-sprint-3.md`](api-sprint-3.md).

## 1) Notifications

### `GET /api/v1/notifications`

Auth. Query: `unread=true`, пагинация `page` / `per_page` как у items.

`200`: `{ "items": [ ... ], "page": 1, "per_page": 20, "total": 0 }`

Элемент: `id`, `item_id`, `to_status`, `title`, `read_at`, `created_at`.

### `POST /api/v1/notifications/{id}/read`

`204`.

### `POST /api/v1/notifications/read-all`

`204`.

## 2) SSE

### `GET /api/v1/events`

Auth: `Authorization: Bearer` или `?access_token=` (EventSource не умеет заголовки).

```text
event: notification
data: {"id":"...","item_id":"...","to_status":"expiring","title":"..."}

event: ping
data: {}
```

`Content-Type: text/event-stream`. Первый кадр — `ping`. Дальше ping каждые 15 с. Hub в памяти процесса.

## 3) Push

### `GET /api/v1/push/vapid-public`

`200`: `{ "public_key": "..." }`

### `POST /api/v1/push/subscribe`

```json
{
  "endpoint": "https://...",
  "keys": { "p256dh": "...", "auth": "..." }
}
```

`204` (upsert по endpoint).

### `DELETE /api/v1/push/subscribe`

`{ "endpoint": "https://..." }` → `204`. Нет строки — тоже `204`.

Тикер после INSERT в `notifications` шлёт тот же JSON, что SSE (`id`, `item_id`, `to_status`, `title`), всем строкам `push_subscriptions`. Ответ `410 Gone` — строку удаляем.

## 4) Dashboard

### `GET /api/v1/dashboard`

```json
{
  "counts": { "active": 0, "expiring_7": 0, "expiring_30": 0, "expired": 0 },
  "upcoming_cost": [{ "currency": "RUB", "monthly": 0, "yearly": 0 }],
  "expirations_by_month": [{ "month": "2026-09", "count": 0 }],
  "cost_by_kind": [{ "kind_id": "...", "currency": "RUB", "amount": 0 }],
  "soonest": []
}
```

`expiring_7` / `expiring_30` — по фактической дате, не только по полю `status`.

`soonest` — до 10 items (краткая карточка: id, title, expires_at, status, kind_id).

`cancelled` / `archived` не входят. `upcoming_cost` — run-rate active/expiring: monthly×12 и yearly/12 (целое), `one_time` не входит. `expirations_by_month` — текущий месяц UTC и следующие 5. `cost_by_kind` — сумма `cost_amount` как записана, без периода.

## 5) Calendar

### `GET /api/v1/calendar?year=2026&month=8`

`200`:

```json
{
  "year": 2026,
  "month": 8,
  "days": [
    { "date": "2026-08-21", "items": [{ "id": "...", "title": "...", "status": "expiring" }] }
  ]
}
```

Пустые дни можно не включать.
