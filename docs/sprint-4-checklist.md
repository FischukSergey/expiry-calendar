# Sprint 4 Checklist

Источник: [`sprint-4-plan.md`](sprint-4-plan.md).

## 1) Статусы и уведомления

- [x] Тикер / ручной tick в тестах
  Примечание: `service.Ticker.Tick` — тот же `StatusAtWrite`. `Run` в `cmd/server` сразу и каждые 60 с; тесты зовут `Tick`.
- [x] Не трогает `cancelled` / `archived`
  Примечание: `ListOpen` / `SetStatus` отсекают эти статусы. `TestTickerSkipsCancelledArchived`.
- [x] Таблица `notifications`, уникальность на день
  Примечание: `009_notifications.sql` — unique `(item_id, to_status, день UTC)`. Повторный tick не плодит строки.
- [x] `GET /api/v1/notifications`
  Примечание: auth, `unread=true`, пагинация как у items. Viewer 200.
- [x] `POST /api/v1/notifications/{id}/read`
  Примечание: 204; повтор уже прочитанного — тоже 204; нет id → 404.
- [x] `POST /api/v1/notifications/read-all`
  Примечание: 204, все непрочитанные. `TestNotificationsReadFlow`.

## 2) SSE

- [x] `GET /api/v1/events` (Bearer или `access_token`)
  Примечание: `BearerOrQuery`. EventSource — `?access_token=`. Без токена 401. nginx `/api/v1/events` без буфера.
- [x] События `notification` и `ping`
  Примечание: сразу ping, дальше каждые 15 с. Тикер после INSERT шлёт notification. `TestEventsSeesTickerNotification`.
- [x] Hub безопасен для горутин
  Примечание: `sse.Hub` mutex; полный буфер клиента не стопорит Publish. `TestHubConcurrent`.

## 3) Web Push

- [x] `GET /api/v1/push/vapid-public`
  Примечание: auth, `{ "public_key": "..." }`. Пустые `VAPID_*` — пара на процесс, в логе warn.
- [x] `POST /api/v1/push/subscribe`
  Примечание: 204, upsert по `endpoint`. Viewer 200. `TestPushSubscribeUpsertAndDelete`.
- [x] `DELETE /api/v1/push/subscribe`
  Примечание: `{ "endpoint" }` → 204; нет строки — тоже 204.
- [x] Рассылка из тикера
  Примечание: `Fanout` после INSERT шлёт SSE и Web Push всем подпискам. Payload как у SSE.
- [x] 410 → удаление подписки
  Примечание: `Broadcast` при 410 удаляет строку. `TestPushTickerBroadcastAnd410`, `TestWebPushSenderSees410`.

## 4) Обзор

- [x] `GET /api/v1/dashboard` (counts, upcoming_cost по валютам, expirations_by_month, cost_by_kind, soonest)
  Примечание: один `ListOpen`. Валюты отдельно. `expiring_7/30` по дате. Viewer 200. `TestDashboardViewerTwoCurrencies`.
- [x] `GET /api/v1/calendar?year=&month=`
  Примечание: пустые дни опущены. year/month обязательны, иначе 422. `TestCalendarQueryAndEmptyMonth`.

## 5) Тесты и DoD

- [x] Integration тикера
  Примечание: `TestTickerIntegrationStatusAndNotification` — Tick → status expiring, GET /notifications, series дашборда. SSE: `TestEventsSeesTickerNotification`.
- [x] Dashboard две валюты без конвертации
  Примечание: `TestDashboardViewerTwoCurrencies` — RUB monthly 200/2400 и USD yearly 36/3 отдельно.
- [x] [`api-sprint-4.md`](api-sprint-4.md)
  Примечание: ручки notifications/events/push/dashboard/calendar совпадают с хендлерами и `openapi.yaml`.
- [x] [`known-limitations-sprint-4.md`](known-limitations-sprint-4.md)
  Примечание: один процесс SSE, нет UI, VAPID на процесс, нет Postgres в `task test:integration`.
