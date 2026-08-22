# Known Limitations — Sprint 1

Осознанные дыры каркаса. Кандидаты в следующие спринты.

## API

**Нет предметного API.** Только `/healthz`. Auth, справочники, записи — Sprint 2–3.

**OpenAPI покрывает только текущие ручки.** Живой `/docs` уже есть; kinds/items и остальное добавим в спеку вместе с хендлерами. Полная сверка — Sprint 6.

## Auth

**Нет пользователей и JWT.** Любой, кто видит порт, видит только health.

## Данные

**Нет прикладных таблиц и seed.** Goose может содержать только служебную миграцию.

## Frontend

**Заглушка, не PWA.** Нет роутера, форм, манифеста, service worker.

## Инфра

**Один инстанс backend.** SSE hub в памяти появится в Sprint 4 и останется однопроцессным в v1.

**Нет HTTPS на локали.** Демо на `http://localhost`. TLS — только `deploy/prod`.

**Хостовый Postgres не на 5432.** Локаль: `localhost:15432`. Прод порт БД не публикует.

**Frontend-образ — SPA-заглушка.** Один экран, без роутера и запросов к `/api/v1`. PWA — Sprint 5.

**CI без integration и compose build.** В Actions только unit-тесты Go, golangci-lint и npm lint/typecheck. `docker compose build` и тесты с Postgres в workflow нет.
