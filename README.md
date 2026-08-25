# Duekeep

Календарь истечений: домены, подписки, аренда, договоры, страховки.

Backend Sprint 1: `GET /healthz`. Swagger: `http://localhost:8080/docs` (через UI — `http://localhost/docs`). Frontend: заглушка на `http://localhost` (nginx проксирует `/api`, `/healthz`, `/docs`).

Демо-аккаунты (только локальный стенд, не прод-секреты):

- `admin@duekeep.local` / `admin1234` — роль admin
- `viewer@duekeep.local` / `viewer1234` — роль viewer

- Функционал: [FUNCTIONAL.md](FUNCTIONAL.md)
- Архитектура: [ARCHITECTURE.md](ARCHITECTURE.md)
- Спринты: [docs/README.md](docs/README.md)
- Журнал работы: [REPORT.md](REPORT.md)

Стек: Go 1.25, React + TypeScript + Vite (PWA), PostgreSQL, JWT + refresh, SSE + Web Push, `docker compose up`.

Команды разработки — [Taskfile.yml](Taskfile.yml) (как в my-chat). Нужен [Task](https://taskfile.dev):

```bash
task --list
task tools:install   # gofumpt, golangci-lint, goose
task fmt
task lint            # Go в Docker + frontend, если есть
task test            # go test -race
```

`task` без аргументов: tidy → fmt → lint → test → build.

Проверка API: `http://localhost:8080/healthz`. Swagger: `http://localhost:8080/docs`. Postgres с хоста: `localhost:15432`.

Локальный стек: `task local:up` / `local:down` (файл [`deploy/local/docker-compose.local.yml`](deploy/local/docker-compose.local.yml)). Из корня для сдачи по-прежнему `docker compose up`.

Прод и секреты: [`deploy/README.md`](deploy/README.md). На VPS — `.env` из [`.env.example`](.env.example), `chmod 600`, `task prod:up`. `.env` в git не класть.

## Статус

Sprint 1 закрыт (каркас). Дальше — Sprint 2 (auth и справочники).
