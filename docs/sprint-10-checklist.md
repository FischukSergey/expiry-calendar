# Sprint 10 Checklist

Источник: [`sprint-10-plan.md`](sprint-10-plan.md). Код — только когда пользователь попросил реализовать спринт. Документы можно завести заранее.

## 1) Схема

- [ ] Миграция `013_item_payments.sql`: таблица, UNIQUE `(item_id, paid_on)`, `owner_id`, CASCADE с item.
- [ ] Backfill: `items.status = paid` → строка на `expires_at` (amount/currency с записи).
- [ ] Статус записи при backfill не затираем.

## 2) API оплаты вхождения

- [ ] `POST /items/{id}/payments` `{ date }`: снимок суммы, идемпотентно (повтор — 200, та же строка).
- [ ] `DELETE /items/{id}/payments?date=`: нет строки — 204.
- [ ] Дата не из ряда записи → 422. Чужой item → 404. Viewer → 403.
- [ ] Audit `pay` / `unpay` в той же транзакции.

## 3) Календарь и обзор

- [ ] `GET /calendar`: у точки `cost_amount`, `currency`, `occurrence_status` (`open`|`paid`). Оплаченный день не прячем.
- [ ] График / `expiring_7`/`30` / `soonest` — только open-вхождения. У soonest дата = ближайшее open (`expires_at`).
- [ ] Карточка: в ответе `next_open_at` (дата или `null`).
- [ ] `upcoming_cost` без изменения формулы Sprint 4.

## 4) Тикер

- [ ] Заморозка `items.status = paid` и `notify_before_days IS NULL` — как Sprint 9.
- [ ] Иначе порог от ближайшего open-вхождения, не сырой `expires_at`, если эта дата уже в `item_payments`.

## 5) UI

- [ ] Сайдбар дня: название, сумма, бейдж вхождения, admin — «Оплачено» / «Снять оплату».
- [ ] Точки сетки по `occurrence_status`. Viewer без кнопок.
- [ ] Карточка и soonest: admin — «Оплатить» ближайшее open (та же ручка).
- [ ] Форму «Статус вручную» не убираем.

## 6) Спека и тесты

- [ ] [`api-sprint-10.md`](api-sprint-10.md) и `backend/openapi.yaml`.
- [ ] Тесты: pay/unpay, идемпотентность, 422/403/404, monthly сентябрь≠октябрь, dashboard без paid-дня, тикер, backfill.

## 7) DoD

- [ ] [`known-limitations-sprint-10.md`](known-limitations-sprint-10.md) заполнен.
- [ ] Демо плана §7 пройдено.
- [ ] `task lint` и `task test` зелёные.
