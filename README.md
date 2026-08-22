# Duekeep

Календарь истечений: домены, подписки, аренда, договоры, страховки.

Backend Sprint 1: `GET /healthz`. Frontend: заглушка на `http://localhost` (nginx проксирует `/api` и `/healthz`).

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

`task` без аргументов: tidy → fmt → lint → test.

Проверка API: `http://localhost:8080/healthz`. Postgres с хоста: `localhost:15432`.

Локальный стек: `task local:up` / `local:down` (файл [`deploy/local/docker-compose.local.yml`](deploy/local/docker-compose.local.yml)). Из корня для сдачи по-прежнему `docker compose up`.

Прод и секреты: [`deploy/README.md`](deploy/README.md). На VPS — `.env` из [`.env.example`](.env.example), `chmod 600`, `task prod:up`. `.env` в git не класть.

## Статус

Sprint 1: разделы 1–4 закрыты. Дальше — CI (раздел 5).
