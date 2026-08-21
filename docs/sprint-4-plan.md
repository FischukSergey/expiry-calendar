# Sprint 4 — статусы, realtime, обзор

Источник: [`ARCHITECTURE.md`](../ARCHITECTURE.md) §6.5, §9, §8.

## 1) Цель

Автоматические статусы, колокольчик, SSE, Web Push, дашборд и календарь — всё ещё API-first (экраны в Sprint 5).

## 2) Входные условия

Sprint 3: items, renew, audit.

## 3) Границы

### Входит

- тикер 60 с + тот же расчёт, что при записи;
- `notifications`, идемпотентность `(item_id, to_status, day)`;
- SSE `/events`, hub в памяти;
- VAPID, subscribe/unsubscribe, рассылка Web Push из тикера;
- `GET /dashboard`, `GET /calendar`;
- суммы дашборда **по валютам отдельно**.

### Не входит

- полноценный UI дашборда/календаря (Sprint 5);
- email/Telegram;
- несколько реплик backend.

## 4) Backlog

### A. Clock + ticker в `cmd/server`, выключается в тестах.

### B. Notifications service + read / read-all.

### C. SSE hub, ping, query `access_token`.

### D. Push: таблица `push_subscriptions`, webpush-go, удаление 410.

### E. Dashboard и calendar queries (без N+1).

### F. Тесты: тикер меняет статус и создаёт notification; dashboard по RUB/USD; subscribe сохраняется.

## 5) Техрешения

- Переход в `expiring`/`expired` → notification + SSE + push всем подписанным пользователям (данные общие).
- EventSource: `?access_token=`.
- VAPID в `.env`, стабильны между рестартами.

## 6) DoD

- подождать тикер или дёрнуть внутренний run once в тесте — статус и уведомление есть;
- вторая «сессия» (тест SSE или ручной EventSource) видит event;
- dashboard отдаёт графиковые series;
- [`api-sprint-4.md`](api-sprint-4.md).

## 7) Демо

Item с `expires_at` сегодня, прогон тикера, `GET /notifications`, curl SSE, `GET /dashboard`.

## 8) Риски

- гонка PATCH и тикера → статус в той же транзакции, unique на notification за день.
- пуши на localhost только Chromium — зафиксировать в limitations.

## 9) Артефакты

- ticker, SSE, push, dashboard/calendar API.
