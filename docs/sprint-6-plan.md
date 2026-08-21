# Sprint 6 — сдача

Источник: критерии задания и [`FUNCTIONAL.md`](../FUNCTIONAL.md) «готово для демо».

## 1) Цель

Преподаватель клонирует репо, делает `docker compose up`, проходит сценарии, открывает Swagger, видит зелёный CI.

## 2) Входные условия

Спринты 1–5 закрыты по чеклистам.

## 3) Границы

### Входит

- `openapi.yaml` + Swagger UI на `http://localhost:8080/docs`;
- ≥ 10 осмысленных тестов (добить пробелы);
- полный seed (50+ items, renewals, audit, unread);
- README: логины, порты, PWA, пуши, `/docs`;
- прогон на чистом томе;
- REPORT по факту проблем спринта (не задним числом пачкой).

### Не входит

- новые фичи (вложения, почта, org_id, iCal).
- если всплыл баг сдачи — фикс, не расширение scope.

## 4) Backlog

### A. Сверить openapi.yaml с суммой api-sprint-1…5.

### B. Покрыть тестами дыры (reuse refresh, CSV dry_run, calendar, push 410).

### C. Seed до объёма FUNCTIONAL, даты от Clock.Today().

### D. README и видео опционально.

### E. `docker compose down -v && docker compose up --build` на чистой машине/томе.

## 5) Техрешения

- OpenAPI — файл, не обязательная кодогенерация.
- Пароли демо только в README.

## 6) DoD

Чеклист демо из FUNCTIONAL (пункты 1–10) проходит.

## 7) Демо (5 мин)

Архитектура слоёв, JWT+refresh, kinds+JSONB, тикер, PWA/push, compose.

## 8) Риски

Самый частый провал задания — compose. Прогон в чистом окружении обязателен.

## 9) Артефакты

- живой `/docs`;
- зелёный CI;
- README;
- закрытый REPORT спринта.
