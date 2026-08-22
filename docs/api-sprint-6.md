# API контракты Sprint 6

Новых прикладных ручек нет. Спринт выравнивает живой OpenAPI с уже зафиксированным.

Сводка документов:

- Sprint 1: [`api-sprint-1.md`](api-sprint-1.md) — `/healthz`
- Sprint 2: [`api-sprint-2.md`](api-sprint-2.md) — auth, kinds, categories
- Sprint 3: [`api-sprint-3.md`](api-sprint-3.md) — items, renew, bulk, audit
- Sprint 4: [`api-sprint-4.md`](api-sprint-4.md) — notifications, SSE, push, dashboard, calendar
- Sprint 5: [`api-sprint-5.md`](api-sprint-5.md) — CSV

## 1) Решения спринта

- Канон для проверяющего: `GET http://localhost:8080/docs` (Swagger UI). UI и `GET /openapi.yaml` подключены сразу после Sprint 1; в этом спринте сверяем спеку со всеми ручками 1–5.
- `backend/openapi.yaml` не имеет права расходиться с markdown выше. При конфликте сначала правим оба, потом код.
- Схема auth в OpenAPI: `bearerAuth`. Для refresh описать cookie и body.

## 2) `GET /docs`

Anon. HTML Swagger UI.

`GET /openapi.yaml` (или встроенный в UI) — сырой spec.

## 3) Совместимость

Запрещено в Sprint 6:

- переименовывать поля без версии `/api/v2`;
- убирать query-параметры списка;
- менять коды ошибок.

Можно: описания, примеры, недостающие 4xx в spec.
