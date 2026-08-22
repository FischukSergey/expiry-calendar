# API контракты Sprint 1

Источник: [`sprint-1-plan.md`](sprint-1-plan.md), [`ARCHITECTURE.md`](../ARCHITECTURE.md) §8.

## 1) Решения спринта

- Публичный API backend — liveness/readiness к БД плюс живой OpenAPI (`/docs`, `/openapi.yaml`).
- Реализация: `handler` → `service.Health` → `repository.Health` (`SELECT 1`).
- Префикс будущего API: `/api/v1`. В Sprint 1 его ручек нет.
- Ошибки сразу в целевом конверте (чтобы не ломать клиентов позже).

## 2) Общие соглашения

- JSON, UUID-строки, время RFC3339.
- Идентификаторы ещё не используются.

Формат ошибки:

```json
{
  "error": {
    "code": "validation_error",
    "message": "...",
    "details": {}
  }
}
```

Коды, которые появятся в следующих спринтах: `unauthorized`, `forbidden`, `not_found`, `conflict`, `validation_error`. В Sprint 1 для `/healthz` тело при 503 допускается простым `{"error":{"code":"internal","message":"database unavailable"}}`.

## 3) Endpoints

### `GET /healthz`

Кто: anon.

Назначение: процесс жив и видит PostgreSQL.

Response `200`:

```json
{
  "status": "ok"
}
```

Response `503`: БД недоступна.

Не требует auth.

### Frontend

`GET /` — HTML-заглушка Vite/React. Не API.

`/api/*`, `/healthz`, `/docs` и `/openapi.yaml` с `:80` проксируются nginx фронта на backend.

### `GET /docs`

Anon. Swagger UI (редирект на `/docs/`).

### `GET /openapi.yaml`

Anon. Сырой OpenAPI 3 документ (`backend/openapi.yaml`, встроен в бинарь).

Исключение: `/healthz` и `/docs` живут на корне `:8080`, не под `/api/v1`.

## 4) Совместимость

Следующий спринт не меняет `/healthz` и `/docs`. Новые прикладные ручки только под `/api/v1`.
