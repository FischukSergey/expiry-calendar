# Sprint 6 Checklist

Источник: [`sprint-6-plan.md`](sprint-6-plan.md), [`FUNCTIONAL.md`](../FUNCTIONAL.md).

## 1) Контракт и тесты

- [ ] `backend/openapi.yaml` совпадает с ручками
- [ ] Swagger UI: `http://localhost:8080/docs`
- [ ] ≥ 10 unit/integration тестов, все осмысленные
- [ ] CI: lint + тесты зелёные

## 2) Seed и запуск

- [ ] 2 пользователя, 9 kinds, ≥ 10 категорий
- [ ] ≥ 50 items: ≥ 5 expired, ≥ 8 expiring в 30 днях
- [ ] ≥ 20 renewals, ≥ 15 audit, unread notifications
- [ ] Повторный `compose up` не дублирует seed
- [ ] `docker compose down -v && docker compose up --build` на чистом томе

## 3) Документы

- [ ] README: запуск, логины, порты, `/docs`, PWA, пуши
- [ ] REPORT дополнен по ходу спринта
- [ ] [`api-sprint-6.md`](api-sprint-6.md) — сводка «без новых полей»

## 4) Демо преподавателя

- [ ] Вход admin и viewer
- [ ] Дашборд с цифрами и графиками
- [ ] Фильтр, карточка, create/edit, renew
- [ ] Календарь
- [ ] CSV import/export
- [ ] Уведомление без перезагрузки
- [ ] PWA / пуш (Chrome)
- [ ] Swagger
- [ ] CI зелёный

## 5) Limitations

- [ ] [`known-limitations-sprint-6.md`](known-limitations-sprint-6.md) = осознанный out of scope v1
