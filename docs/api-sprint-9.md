# API контракты Sprint 9

Источник: [`sprint-9-plan.md`](sprint-9-plan.md). База — сумма [`api-sprint-1.md`](api-sprint-1.md)…[`api-sprint-8.md`](api-sprint-8.md). Новых путей нет. Меняются `status` и `notify_before_days`. Справочник kinds пополняется seed-строкой.

## 1) Решения

- `paid` — ещё одно значение `items.status`, только вручную. Пока статус `paid`, уведомлений нет (предоплата до `expires_at`).
- «Не уведомлять» — `notify_before_days: null`, не отдельное поле и не `-1`.
- Тип «Мобильная связь» — строка `item_kinds` (`slug=mobile`), не новая ручка и не UI создания kinds.

## 2) Item

Поля как Sprint 3, отличия:

### `status`

Было: `active` | `expiring` | `expired` | `cancelled` | `archived`.  
Стало: плюс **`paid`**.

- Create/PATCH/bulk: клиент может прислать `paid` (как `cancelled` / `archived`).
- Иное неизвестное значение → `422 validation_error`.
- Тикер не меняет `paid` и не создаёт по такой записи notification / SSE / push, даже если `notify_before_days` число.
- `notify_before_days` при `paid` не затирается: после `renew` или снятия ручного статуса порог снова используется.
- Фильтр `GET /items?status=paid` валиден.
- Дашборд / календарь открытых: `paid` не входит в «сгорит» / ближайшие сроки.

### `notify_before_days`

- Тип JSON: целое `≥ 0` **или** `null`.
- `null` — не уведомлять: тикер не создаёт notification и не ставит `expiring`.
- `0` — порог в календарный день `expires_at` (как сейчас).
- Опущенное поле на create — по-прежнему дефолт `30`, не `null`.
- PATCH: явный `null` сбрасывает порог; отсутствие ключа поле не меняет.

Пример PATCH:

```json
{ "status": "paid" }
```

```json
{ "notify_before_days": null }
```

Тело Item в ответах: `notify_before_days` может быть `null`; `status` может быть `"paid"`. `owner_id` по-прежнему не в JSON.

## 3) Kinds

Ручки как [`api-sprint-2.md`](api-sprint-2.md), без новых путей и без продуктового UI create.

`GET /kinds` после `EnsureKinds` / seed содержит среди прочих:

```json
{ "slug": "mobile", "name": "Мобильная связь" }
```

`id`, `color`, `attr_schema` — как у остальных kind. Справочник общий на инсталляцию.

## 4) CSV

- Колонка `status` принимает `paid`.
- Колонка `notify_before_days`: число как раньше; пусто, `off`, `-` → `null`.
- `kind` / slug может быть `mobile`.

## 5) Совместимость

- Старые клиенты, которые всегда шлют число дней, работают.
- Клиент, который считает `status` закрытым enum без `paid`, должен принять новое значение в списке — ломающее чтение, сознательно.
- `/healthz`, auth, JWT не меняем.
- OpenAPI править вместе с хендлерами.
