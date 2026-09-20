# Duekeep

Календарь обязательств: домены, подписки, аренда, договоры, страховки, налоги, ТО.

Преподаватель клонирует репозиторий и поднимает стек одной командой. Текущая защита — ветка `main` (свои данные, оплата вхождения, CD).

## Запуск

Нужны Docker Engine и Docker Compose v2 (`include` в корневом файле). Порты `80`, `8080`, `15432`.

```bash
git clone https://github.com/FischukSergey/expiry-calendar.git
cd expiry-calendar
docker compose up --build
```

Уже в корне, если нужен чистый том: `docker compose down -v && docker compose up --build`.

Три сервиса: PostgreSQL, backend (Go), frontend (nginx + SPA). Backend ждёт healthy у БД, накатывает goose и пишет локальный seed (`SEED=true` в local compose). Корневой `.env` не нужен и не читается.

Повторный `docker compose up` не дублирует пользователей, виды, категории, записи и оплаты: конфликт по стабильным id / email / slug.

Локальный `go run` без `SEED=true` демо не пишет. Compose и `task local:up` ставят флаг сами.

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
| `viewer@duekeep.local` | `viewer1234` | чтение своего пустого списка, без кнопок записи |

Каталог 50+ принадлежит seed-admin: типы включая «Мобильная связь», запись «не уведомлять», заморозка `paid`, оплаты вхождений на календаре. Viewer чужие записи не видит. На проде seed выключен (`SEED=false` в prod compose, не из `.env`): нет этих аккаунтов и нет демо-записей.

## Сценарий демо

1. Войти admin (подсказка на `/login` только локально), затем viewer (кнопки записи скрыты).
2. Дашборд: KPI, суммы оплаты по месяцам, pie по валюте, топ-10.
3. Список: фильтр, карточка, создать/править, продлить (история на карточке). Тип «Мобильная связь», статус «Оплачено», чекбокс «Не уведомлять».
4. Календарь: дни с бейджем оплаты и открытые вхождения; «Оплатить» в сайдбаре дня, на карточке и в soonest.
5. Экспорт CSV фильтра; импорт — dry run, затем запись.
6. Колокольчик: непрочитанные; вторая вкладка — SSE без перезагрузки (смена срока у записи; тикер при старте и каждые 12 ч).
7. Профиль: «Установить» (Chrome), разрешение пушей.
8. Swagger: `/docs`.
9. CI: вкладка Actions, workflow `CI`. Прод `duekeep.ru` обновляется с `main` после зелёного CI ([deploy/README.md](deploy/README.md)).

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

На VPS (`duekeep.ru`, `159.194.252.6`) прод обновляется с `main`: после lint/test/build/frontend GitHub Actions по SSH запускает [`deploy/prod/deploy.sh`](deploy/prod/deploy.sh). Секреты приложения только в `.env` на сервере. Подробности и откат: [deploy/README.md](deploy/README.md). `.env` в git не класть.

## Статус

Текущая защита — `main`. Историческая сдача v1 — тег [`v1.0.0`](https://github.com/FischukSergey/expiry-calendar/releases/tag/v1.0.0) (Sprint 6, общий каталог). Прод: каждый видит своё ([Sprint 7](docs/sprint-7-plan.md)); CD с `main` — [Sprint 8](docs/sprint-8-plan.md). Продукт: [Sprint 9](docs/sprint-9-plan.md) (оплачено, «Мобильная связь», «не уведомлять»), [Sprint 10](docs/sprint-10-plan.md) (оплата вхождения).
