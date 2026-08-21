# API контракты Sprint 5

Новое — только CSV. Остальное: [`api-sprint-4.md`](api-sprint-4.md).

## 1) Export

### `GET /api/v1/items/export`

Auth (viewer и admin). Те же query, что `GET /items`, без пагинации (или с разумным max, например 10_000).

`200` `text/csv`. Колонки: `id`, `title`, `kind_slug`, `status`, `expires_at`, `cost_amount`, `currency`, `vendor`, `billing_period`, `category_name`, `tags`, плюс известные `attrs.*` из schema видов в выгрузке.

## 2) Import

### `POST /api/v1/items/import?dry_run=true`

Admin. `multipart/form-data`:

- `file` — CSV
- `mapping` — JSON: `{ "title": "Name", "kind_slug": "Type", "expires_at": "Until", "attrs.registrar": "Reg" }`

`200`:

```json
{
  "rows": 10,
  "valid": 8,
  "errors": [{ "line": 3, "message": "expires_at required" }],
  "preview": []
}
```

Не пишет в БД.

### `POST /api/v1/items/import`

Тот же multipart без dry_run или `dry_run=false`. Транзакция, audit `items.import`.

`200`: `{ "created": 8 }` или `422` если есть ошибки валидации (ничего не пишем).

## 3) UI не является API

Экраны используют уже описанные ручки. Новые поля на wire — только через правку sprint api.
