# API контракты Sprint 1

Источник: [`sprint-1-plan.md`](sprint-1-plan.md), [`ARCHITECTURE.md`](../ARCHITECTURE.md) §8.

## 1) Решения спринта

- Единственный публичный API backend — liveness/readiness к БД.
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

`GET /` — HTML-заглушка. Не API.

`/api/*` проксируется на backend. В Sprint 1 кроме `/healthz` (на корне backend, не под `/api/v1`) других маршрутов нет.

Исключение: `/healthz` живёт на корне `:8080`, не под `/api/v1` — так проще проверять compose без префикса.

## 4) Совместимость

Следующий спринт не меняет `/healthz`. Новые ручки только под `/api/v1`.
