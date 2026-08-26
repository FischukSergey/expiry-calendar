# API контракты Sprint 3

Auth и справочники — [`api-sprint-2.md`](api-sprint-2.md).

## 1) Решения

- Item: колонки + `attrs` JSONB.
- Пагинация offset: `page`, `per_page` (max 100), `total`.
- Статус пересчитывается при записи, если не `cancelled`/`archived`.
- `cost_amount` и `new_cost` — целые (≥ 0), без дробной части.

## 2) Item

Поля: `id`, `title`, `description`, `kind_id`, `category_id`, `vendor`, `tags`, `cost_amount`, `currency`, `billing_period` (`one_time`|`monthly`|`yearly`), `started_at`, `expires_at`, `notify_before_days`, `url`, `account_hint`, `status`, `attrs`, `created_at`, `updated_at`.

Обязательны при создании: `title`, `kind_id`, `expires_at`. `cost_amount` — целое ≥ 0. `started_at` ≤ `expires_at`, если обе заданы.

### `GET /api/v1/items`

Query: `q`, `kind_id`, `status`, `category_id`, `vendor`, `expires_from`, `expires_to`, `cost_from`, `cost_to`, `billing_period`, `tag`, `sort` (`expires_at`|`cost_amount`|`title`|`updated_at`), `order` (`asc`|`desc`), `page`, `per_page`.

`200`: `{ "items": [ ... ], "page": 1, "per_page": 20, "total": 50 }`

### `POST /api/v1/items`

Admin. `201` созданный item (status уже посчитан).

### `GET /api/v1/items/{id}`

Auth. Карточка + можно не включать renewals (отдельный список в теле):

```json
{
  "item": {},
  "renewals": []
}
```

### `PATCH /api/v1/items/{id}`

Admin. `200`.

### `DELETE /api/v1/items/{id}`

Admin. `204`.

### `POST /api/v1/items/{id}/renew`

Admin.

```json
{
  "new_expires_at": "2027-08-01",
  "new_cost": 1990,
  "comment": "продлил на год"
}
```

`200` обновлённый item. Пишет `renewals` и audit.

### `POST /api/v1/items/bulk`

Admin. `{ "ids": ["..."], "category_id": "...", "status": "archived" }` — хотя бы одно из `category_id`/`status`. `200`: `{ "updated": 3 }`.

## 3) Audit

### `GET /api/v1/audit`

Admin. Пагинация как у items. Элемент: `id`, `actor_id`, `action`, `entity`, `entity_id`, `before_json`, `after_json`, `created_at`.

## 4) Совместимость

Dashboard/calendar/CSV не ломают эти поля. Новые query-параметры — только с правкой файла.
