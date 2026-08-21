# Sprint 1 Checklist

Источник: [`sprint-1-plan.md`](sprint-1-plan.md).

## 1) Подготовка

- [ ] Структура каталогов backend/frontend по архитектуре.
- [ ] `.env.example` без секретов.
- [ ] Контракт [`api-sprint-1.md`](api-sprint-1.md) не расходится с хендлером.

## 2) Docker

- [ ] `docker-compose.yml`: db, backend, frontend.
- [ ] Healthcheck PostgreSQL.
- [ ] Backend ждёт healthy db.
- [ ] Прогон `docker compose down -v && docker compose up --build`.

## 3) Backend

- [ ] Go 1.25 в `go.mod` и образе.
- [ ] Пул pgx и goose при старте.
- [ ] `GET /healthz` (200 / 503).
- [ ] slog JSON, без секретов.
- [ ] Пакеты `handler` / `service` / `repository` / `model` заведены.

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
