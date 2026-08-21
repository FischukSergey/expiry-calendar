# Sprint 4 Checklist

Источник: [`sprint-4-plan.md`](sprint-4-plan.md).

## 1) Статусы и уведомления

- [ ] Тикер / ручной tick в тестах
- [ ] Не трогает `cancelled` / `archived`
- [ ] Таблица `notifications`, уникальность на день
- [ ] `GET /api/v1/notifications`
- [ ] `POST /api/v1/notifications/{id}/read`
- [ ] `POST /api/v1/notifications/read-all`

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
