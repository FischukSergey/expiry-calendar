# Known Limitations — Sprint 2

## Данные

**Нет записей истечения.** Справочники пустые с точки зрения бизнеса.

**Нет audit_log.** Появится вместе с мутациями items в Sprint 3.

## Auth

**Access до `exp` жив после logout.** Инвалидируется только refresh. Access короткий (15 мин) — принято.

**Нет смены роли в UI/API.** Admin только из seed.

**Нет `org_id` в claims.** Заложено в архитектуре на потом.

**Refresh: body важнее cookie.** Пустой body + cookie — берём cookie.

**Logout без refresh (только access)** не находит текущую сессию в БД и отвечает 204: access не содержит id refresh.

**JWT_SECRET** в local compose короче 32 символов — порог длины не вводим.

## UI

**Нет экрана логина.** Проверка API тестами и curl.

## OpenAPI

Спека `/openapi.yaml` дополняется вместе с новыми ручками; полная сверка — Sprint 6.
