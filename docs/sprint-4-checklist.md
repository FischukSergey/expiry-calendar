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

- [ ] `GET /api/v1/events` (Bearer или `access_token`)
- [ ] События `notification` и `ping`
- [ ] Hub безопасен для горутин

## 3) Web Push

- [ ] `GET /api/v1/push/vapid-public`
- [ ] `POST /api/v1/push/subscribe`
- [ ] `DELETE /api/v1/push/subscribe`
- [ ] Рассылка из тикера
- [ ] 410 → удаление подписки

## 4) Обзор

- [ ] `GET /api/v1/dashboard` (counts, upcoming_cost по валютам, expirations_by_month, cost_by_kind, soonest)
- [ ] `GET /api/v1/calendar?year=&month=`

## 5) Тесты и DoD

- [ ] Integration тикера
- [ ] Dashboard две валюты без конвертации
- [ ] [`api-sprint-4.md`](api-sprint-4.md)
- [ ] [`known-limitations-sprint-4.md`](known-limitations-sprint-4.md)
