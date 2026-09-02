# API контракты Sprint 10

Источник: [`sprint-10-plan.md`](sprint-10-plan.md). База — сумма [`api-sprint-1.md`](api-sprint-1.md)…[`api-sprint-9.md`](api-sprint-9.md). Новые пути — только платежи вхождения. Календарь расширяет элемент дня.

## 1) Решения

- Оплата цикла — строка `item_payments` на календарную дату, не `items.status`.
- `items.status = paid` (Sprint 9) не отменяем: заморозка всей записи, форма / PATCH / bulk.
- В календаре статус **дня**: `occurrence_status` = `open` | `paid`. Оплаченный день остаётся в ответе.
- `owner_id` в JSON нет.

## 2) Payments

Auth. Мутации — admin (`RequireAdmin`). Чтение отдельных платежей в этом спринте не заводим: факт виден в `GET /calendar`.

### `POST /api/v1/items/{id}/payments`

```json
{ "date": "2026-09-15" }
```

- `date` — `YYYY-MM-DD`, должна быть вхождением записи (якорь `expires_at`, clamp 29–31, как развёртка Sprint 9). Иначе `422 validation_error`.
- `amount` / `currency` сервер копирует с item. С клиента не принимаем.
- Первая отметка: `201` и тело платежа. Повтор на ту же дату: `200`, та же строка (не 409).
- Нет записи / чужой UUID: `404`.
- Viewer: `403`.

```json
{
  "id": "…",
  "item_id": "…",
  "date": "2026-09-15",
  "amount": 799,
  "currency": "RUB"
}
```

Поле `status` в теле не нужно: наличие строки = оплачено.

### `DELETE /api/v1/items/{id}/payments?date=2026-09-15`

`204`. Нет строки — тоже `204`. Нет item / чужой — `404`. Viewer — `403`. `date` обязателен, иначе `422`.

## 3) Calendar

`GET /api/v1/calendar?year=&month=` — путь Sprint 4. Элемент дня:

```json
{
  "id": "…",
  "title": "Netflix",
  "status": "active",
  "occurrence_status": "open",
  "cost_amount": 799,
  "currency": "RUB"
}
```

- `status` — статус **записи** (как сейчас).
- `occurrence_status` — `paid`, если есть `item_payments` на этот день, иначе `open`.
- `cost_amount` / `currency`: у `paid` — снимок платежа; у `open` — текущие поля item.
- Оплаченное вхождение **входит** в день. `cancelled` / `archived` по-прежнему нет.

Ломающее чтение: старый клиент без новых полей JSON обычно игнорирует лишние ключи; кто закрыл enum элемента — должен принять поля.

## 4) Dashboard

Путь тот же. Смысл:

- `expirations_by_month`, `expiring_7` / `30`, `soonest` — вхождения с `occurrence_status=open`.
- `soonest[].expires_at` — дата **ближайшего open** (её же шлёт клиент в `POST …/payments` с обзора).
- `upcoming_cost` — без изменения Sprint 4.
- Запись с `items.status=paid` (заморозка): в «сгорит» / open-график не входит (как Sprint 9 для всей строки). Её оплаченные дни в календаре всё равно видны, если есть развёртка и/или backfill.

## 5) Item и тикер

Контракт Item (create/PATCH/`paid`/`notify_before_days`) — [`api-sprint-9.md`](api-sprint-9.md).

`GET /items/{id}` (карточка): плюс **`next_open_at`**: дата ближайшего open-вхождения или `null` (заморозка `paid` / нет open). Кнопка на карточке шлёт эту дату в `POST …/payments`.

Тикер: при заморозке `paid` или `notify_before_days: null` — как Sprint 9. Иначе дата для `expiring`/`expired` — ближайшее open-вхождение.

## 6) Совместимость

- `/healthz`, auth, JWT, renew, CSV записей — без обязательных правок.
- OpenAPI править вместе с хендлерами.
