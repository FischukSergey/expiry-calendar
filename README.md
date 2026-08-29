# Duekeep

Календарь истечений: домены, подписки, аренда, договоры, страховки, налоги, ТО.

Преподаватель клонирует репозиторий и поднимает стек одной командой.

## Запуск

Нужны Docker и Docker Compose. Из корня репозитория:

```bash
docker compose down -v && docker compose up --build
```

Или короче, если тома уже не важны: `docker compose up --build`.

Три сервиса: PostgreSQL, backend (Go), frontend (nginx + SPA). Backend ждёт healthy у БД, накатывает goose и пишет seed.

Повторный `docker compose up` не дублирует пользователей, виды, категории и записи: конфликт по стабильным id / email / slug.

Разработка: `task local:up` / `local:down` (тот же проект `duekeep`, файл [`deploy/local/docker-compose.local.yml`](deploy/local/docker-compose.local.yml)).

## Адреса

| Что | URL |
|---|---|
| UI | http://localhost |
| Swagger | http://localhost/docs или http://localhost:8080/docs |
| OpenAPI YAML | http://localhost:8080/openapi.yaml |
| Health | http://localhost/healthz или `:8080/healthz` |
| API | http://localhost/api/v1/… (nginx) или `:8080/api/v1/…` |
| Postgres с хоста | `localhost:15432` |

nginx на `:80` проксирует `/api`, `/healthz`, `/docs`, `/openapi.yaml` на backend.

## Демо-аккаунты

Только локальный стенд, не прод-секреты.

| Email | Пароль | Роль |
|---|---|---|
| `admin@duekeep.local` | `admin1234` | полный CRUD, аудит, импорт |
| `viewer@duekeep.local` | `viewer1234` | чтение, экспорт, без кнопок записи |

Данные v1 общие: оба видят один каталог.

## Сценарий демо

1. Войти admin, затем viewer (кнопки записи скрыты).
2. Дашборд: KPI, столбцы на 6 месяцев, pie по валюте, топ-10.
3. Список: фильтр, карточка, создать/править, продлить (история на карточке).
4. Календарь текущего месяца.
5. Экспорт CSV фильтра; импорт — dry run, затем запись.
6. Колокольчик: непрочитанные из seed; вторая вкладка — SSE без перезагрузки (смена срока у записи и тикер до 60 с).
7. Профиль: «Установить» (Chrome), разрешение пушей.
8. Swagger: `/docs`.
9. CI: вкладка Actions, workflow `CI`.

## PWA и пуши

- Manifest Duekeep, standalone, иконки 192/512, `offline.html`.
- Service worker: HTML/API network-first, SSE не кэшируется, `sw.js` без кэша.
- Установка: Chrome на localhost, часто со второго захода (`beforeinstallprompt`).
- Web Push: после входа браузер спросит разрешение. В local compose VAPID зафиксирован (подписки переживают рестарт). Ориентир — Chromium. Safari/iOS не демо.
- Если `VAPID_*` пустые (прод без `.env`), backend генерирует ключи на процесс.

## Разработка

Нужен [Task](https://taskfile.dev):

```bash
task --list
task tools:install   # gofumpt, golangci-lint, goose
task fmt
task lint            # Go в Docker + frontend
task test            # go test -race
```

`task` без аргументов: tidy → fmt → lint → test → build.

Стек: Go 1.25, React + TypeScript + Vite (PWA), PostgreSQL, JWT + refresh, SSE + Web Push.

- Функционал: [FUNCTIONAL.md](FUNCTIONAL.md)
- Архитектура: [ARCHITECTURE.md](ARCHITECTURE.md)
- Спринты: [docs/README.md](docs/README.md)
- Журнал: [REPORT.md](REPORT.md)
- Прод и секреты: [deploy/README.md](deploy/README.md)

На VPS — `.env` из [`.env.example`](.env.example), `chmod 600`, `task prod:up`. `.env` в git не класть.

## Статус

Sprint 6 — сдача v1. После v1 изоляция данных: [Sprint 7](docs/sprint-7-plan.md).
