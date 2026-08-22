# Sprint 1 — каркас

Источник: [`ARCHITECTURE.md`](../ARCHITECTURE.md) §3–5, §15; [`FUNCTIONAL.md`](../FUNCTIONAL.md).

## 1) Цель спринта

Поднять пустой контур, который преподаватель сможет запустить одной командой. Без бизнес-логики.

К концу спринта:

- `docker compose up --build` поднимает PostgreSQL, backend Go 1.25, frontend-заглушку;
- backend отвечает на `GET /healthz` (пинг БД);
- goose прогоняет пустую/служебную миграцию;
- frontend (Vite + React + Tailwind) открывается на `:80`, nginx проксирует `/api` на backend;
- GitHub Actions: lint + пустые/дымовые тесты — зелёные.

## 2) Границы

### Входит

- репозиторная структура `backend/`, `frontend/` по слоям handler / service / repository / model;
- Dockerfile, compose, `.env.example`;
- chi, `/healthz`, подключение pgx, goose;
- Vite-заглушка («Duekeep, скоро») и nginx;
- CI workflow.

### Не входит

- users, JWT, виды, категории, записи;
- seed данных;
- OpenAPI UI;
- PWA и пуши.

## 3) Backlog

### A. Репозиторий и DX

- `.gitignore`, `.env.example` (`DATABASE_URL`, `HTTP_ADDR`, заготовки `JWT_*`, `VAPID_*`).
- `docker-compose.yml`: `db` (postgres:16, healthcheck, volume), `backend`, `frontend`.
- README: как поднять каркас (порты 80 и 8080).
- Taskfile уже в корне: `fmt` / `lint` / `test` / `tools:install`. В спринте дописать `local:up` / `local:down` и `test:integration`.

### B. Backend

- `cmd/server/main.go`: конфиг из env, slog JSON, graceful shutdown.
- `internal/db`: пул pgx.
- goose: первая миграция (например `schema_migrations` / no-op `001_init.sql` с комментарием).
- `GET /healthz` → 200, если `SELECT 1` ок, иначе 503.
- слои-заготовки пакетов без хендлеров предметной области.

### C. Frontend

- Vite + React 18 + TypeScript + Tailwind.
- одна страница-заглушка.
- `nginx.conf`: SPA fallback, `location /api` → `backend:8080`.

### D. CI

- `go test ./...`, `golangci-lint`.
- `npm ci && npm run lint && npm run typecheck`.

## 4) Техрешения спринта

- Go **1.25**, слои классические, не доменные пакеты.
- Backend снаружи `:8080`, UI `:80`.
- Миграции при старте контейнера (goose), не руками.
- Образ БД: `postgres:16-alpine`.
- Каталог `deploy/` как в my-chat. Корень `docker-compose.yml` — `include` + `name: duekeep`.
- Локальный Postgres на хосте: **15432**. Внутри compose — `db:5432`.
- Прод: секреты только в `.env` на VPS.
- Taskfile `local:*` с проектом `duekeep`, leftover `expiry-calendar` снимается.

## 5) DoD

- чистый `docker compose down -v && docker compose up --build` поднимает три сервиса;
- `curl localhost:8080/healthz` → 200;
- UI открывается в браузере;
- CI зелёный;
- контракт — [`api-sprint-1.md`](api-sprint-1.md).

## 6) Демо

1. `docker compose up --build`
2. Открыть `http://localhost` — заглушка.
3. Открыть `http://localhost:8080/healthz`.
4. Показать зелёный Actions (или локальный lint/test).

## 7) Риски

- compose не дожидается Postgres → healthcheck + `depends_on: condition: service_healthy`.
- frontend ходит на другой origin → только nginx `/api`, не хардкод `:8080` в браузере.

## 8) Артефакты

- compose + Dockerfile;
- пустой сервер и заглушка UI;
- CI;
- docs спринта.
