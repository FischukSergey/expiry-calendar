# Known Limitations — Sprint 2

## Данные

**Нет записей истечения.** Справочники пустые с точки зрения бизнеса.

**Нет audit_log.** Появится вместе с мутациями items в Sprint 3.

## Auth

**Access до `exp` жив после logout.** Инвалидируется только refresh. Access короткий (15 мин) — принято.

**Нет смены роли в UI/API.** Admin только из seed.

**Нет `org_id` в claims.** Sprint 7 тоже не добавляет: изоляция по `owner_id`, не по org.

**Refresh: body важнее cookie.** Пустой body + cookie — берём cookie.

**Logout без refresh (только access)** не находит текущую сессию в БД и отвечает 204: access не содержит id refresh.

**JWT_SECRET** в local compose короче 32 символов — порог длины не вводим.

## Справочники

**`CountItems` всегда 0.** Таблицы `items` нет до Sprint 3 — DELETE kind/category по «занятости» записями не сработает.

**Валидация `attrs` записи против `attr_schema`** — Sprint 3. Здесь проверяется только форма самого schema.

**GET `/categories`** — дерево в `{items:[корни]}`, не плоский список.

## UI

**Нет экрана логина.** Проверка API тестами и curl.

## Тесты

**Нет integration с Postgres.** Сценарии login→refresh→logout и viewer 403 — httptest + память. `task test:integration` и CI с БД по-прежнему заглушка.

## OpenAPI

Спека `/openapi.yaml` дополняется вместе с новыми ручками; полная сверка — Sprint 6.
