# Sprint 10 Checklist

Источник: [`sprint-10-plan.md`](sprint-10-plan.md). Код — только когда пользователь попросил реализовать спринт. Документы можно завести заранее.

## 1) Схема

- [x] Миграция `013_item_payments.sql`: таблица, UNIQUE `(item_id, paid_on)`, `owner_id`, CASCADE с item.
- [x] Backfill: `items.status = paid` → строка на `expires_at` (amount/currency с записи).
- [x] Статус записи при backfill не затираем.

## 2) API оплаты вхождения

- [x] `POST /items/{id}/payments` `{ date }`: снимок суммы, идемпотентно (повтор — 200, та же строка).
- [x] `DELETE /items/{id}/payments?date=`: нет строки — 204.
- [x] Дата не из ряда записи → 422. Чужой item → 404. Viewer → 403.
- [x] Audit `pay` / `unpay` в той же транзакции.

## 3) Календарь и обзор

- [x] `GET /calendar`: у точки `cost_amount`, `currency`, `occurrence_status` (`open`|`paid`). Оплаченный день не прячем.
- [x] График / `expiring_7`/`30` / `soonest` — только open-вхождения. У soonest дата = ближайшее open (`expires_at`).
- [x] Карточка: в ответе `next_open_at` (дата или `null`).
- [x] `upcoming_cost` без изменения формулы Sprint 4.

## 4) Тикер

- [x] Заморозка `items.status = paid` и `notify_before_days IS NULL` — как Sprint 9.
- [x] Иначе порог от ближайшего open-вхождения, не сырой `expires_at`, если эта дата уже в `item_payments`.

## 5) UI

- [x] Сайдбар дня: название, сумма, бейдж вхождения, admin — «Оплачено» / «Снять оплату».
- [x] Точки сетки по `occurrence_status`. Viewer без кнопок.
- [x] Карточка и soonest: admin — «Оплатить» ближайшее open (та же ручка).
- [x] Форму «Статус вручную» не убираем.

## 6) Спека и тесты

- [x] [`api-sprint-10.md`](api-sprint-10.md) и `backend/openapi.yaml`.
- [x] Тесты: pay/unpay, идемпотентность, 422/403/404, monthly сентябрь≠октябрь, dashboard без paid-дня, тикер, backfill.

## 7) DoD

- [x] [`known-limitations-sprint-10.md`](known-limitations-sprint-10.md) заполнен.
- [x] Демо плана §7: API на локальной Postgres (миграция 013, backfill 1=1, monthly сентябрь paid / октябрь open, next_open_at, идемпотентный 200, unpay, 422/403). UI в бандле; compose-backend не пересобран (Docker Hub timeout).
- [x] `task lint` и `task test` зелёные.
