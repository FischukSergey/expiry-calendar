# Sprint 1 Checklist

Источник: [`sprint-1-plan.md`](sprint-1-plan.md).

## 1) Подготовка

- [x] Структура каталогов backend/frontend по архитектуре.
  Примечание: пакеты `internal/{handler,service,repository,model,middleware,clock,sse,seed,db}`, `cmd/server`, `migrations`, `testdata`; frontend `src/{pages,components,api,hooks}`. Кода ручек и Vite ещё нет — разделы 3–4.
- [x] `.env.example` без секретов.
  Примечание: демо-значения `JWT_SECRET=dev-only-change-me`, пустые VAPID. `.env` в gitignore.
- [x] Контракт [`api-sprint-1.md`](api-sprint-1.md) не расходится с хендлером.
  Примечание: сверка в §3 — `GET /healthz` → `{"status":"ok"}` / 503 `internal`.

## 2) Docker

- [x] `docker-compose.yml`: db, backend, frontend.
  Примечание: стек перенесён в `deploy/local/`; корневой `docker-compose.yml` делает `include` (сдача одной командой из корня). Прод/секреты — `deploy/prod/` + корневой `.env.example`.
- [x] Healthcheck PostgreSQL.
  Примечание: `pg_isready -U duekeep -d duekeep`. Образ `postgres:16-alpine` (локальный кэш; pull `postgres:16` с Docker Hub отвалился по таймауту).
- [x] Backend ждёт healthy db.
  Примечание: `depends_on.db.condition: service_healthy`.
- [x] Прогон `docker compose down -v && docker compose up --build`.
  Примечание: три сервиса Up, db healthy, frontend `:80` → 200, backend `:8080` проброшен. Образы backend/frontend пока заглушки (разделы 3–4). Порт 8080 был занят `my-chat-main-service-local` — контейнер остановлен на время прогона.

## 3) Backend

- [x] Go 1.25 в `go.mod` и образе.
  Примечание: `go 1.25.7`, образ `golang:1.25-alpine`.
- [x] Пул pgx и goose при старте.
  Примечание: `001_init.sql` (no-op `SELECT 1`), embed в `migrations.FS`.
- [x] `GET /healthz` (200 / 503).
  Примечание: в контейнере `curl :8080/healthz` → `{"status":"ok"}` 200. Юнит-тесты ok/503.
- [x] slog JSON, без секретов.
  Примечание: в логах `database_url` с паролем `***`.
- [x] Пакеты `handler` / `service` / `repository` / `model` заведены.
  Примечание: healthz идёт по слоям; остальные пакеты — заготовки.

## 4) Frontend

- [ ] Vite + React + TS + Tailwind.
- [ ] Заглушка на `/`.
- [ ] nginx: статика + `/api` proxy.

## 5) CI и качество

- [ ] GitHub Actions: go test + golangci-lint.
- [ ] npm lint + typecheck.
- [ ] Хотя бы один дымовой `TestHealthz` или `TestMain` package.

## 6) DoD

- [ ] Три контейнера поднимаются одной командой.
- [ ] `/healthz` зелёный.
- [ ] Зафиксированы [`known-limitations-sprint-1.md`](known-limitations-sprint-1.md).
