# Sprint 6 Checklist

Источник: [`sprint-6-plan.md`](sprint-6-plan.md), [`FUNCTIONAL.md`](../FUNCTIONAL.md).

## 1) Контракт и тесты

- [x] `backend/openapi.yaml` совпадает с ручками
  Примечание: пути 1–5 в spec; cookie `duekeep_refresh` на refresh/logout; `bearerAuth`. `TestOpenAPISpec`.
- [x] Swagger UI: `http://localhost:8080/docs`
  Примечание: `:8080/docs/` и `localhost/docs/` — 200, HTML swagger. `TestDocsUI`.
- [x] ≥ 10 unit/integration тестов, все осмысленные
  Примечание: 83 теста. Дыры плана уже были: reuse refresh, CSV dry_run, calendar, push 410.
- [x] CI: lint + тесты зелёные
  Примечание: `task lint` / `task test` зелёные. Actions — тот же workflow; зелёный run после пуша.

## 2) Seed и запуск

- [x] 2 пользователя, 9 kinds, ≥ 10 категорий
  Примечание: 2 / 9 / 13. `CheckCatalog`.
- [x] ≥ 50 items: ≥ 5 expired, ≥ 8 expiring в 30 днях
  Примечание: 52 items, 6 expired, 12 expiring. Даты от `Clock.Today()`.
- [x] ≥ 20 renewals, ≥ 15 audit, unread notifications
  Примечание: 22 / 24 / 18 unread.
- [x] Повторный `compose up` не дублирует seed
  Примечание: после restart те же 52 / 18 / 24.
- [x] `docker compose down -v && docker compose up --build` на чистом томе
  Примечание: три сервиса healthy, новый том, seed накатился.

## 3) Документы

- [x] README: запуск, логины, порты, `/docs`, PWA, пуши
- [x] REPORT дополнен по ходу спринта
- [x] [`api-sprint-6.md`](api-sprint-6.md) — без переименований; у `expirations_by_month` добавлен `amounts`

## 4) Демо преподавателя

- [x] Вход admin и viewer
  Примечание: login 200; viewer GET items 200, POST 403.
- [x] Дашборд с цифрами и графиками
  Примечание: counts ненулевые, upcoming_cost RUB+USD. Столбцы — суммы оплаты по месяцам (`amounts`).
- [x] Фильтр, карточка, create/edit, renew
  Примечание: API smoke create/patch/renew/delete; карточка с renewals.
- [x] Календарь
  Примечание: текущий месяц, 5 дней с отметками.
- [x] CSV import/export
  Примечание: export текущего фильтра 200; import — тесты dry_run/запись Sprint 5.
- [x] Уведомление без перезагрузки
  Примечание: 18 unread из seed; SSE/инвалидация — Sprint 5 (`TestEventsSeesTickerNotification`).
- [x] PWA / пуш (Chrome)
  Примечание: manifest, SW, иконки, offline — Sprint 5. Демо в Chrome на localhost.
- [x] Swagger
- [x] CI зелёный
  Примечание: локально lint+test; Actions после пуша.

## 5) Limitations

- [x] [`known-limitations-sprint-6.md`](known-limitations-sprint-6.md) = осознанный out of scope v1
