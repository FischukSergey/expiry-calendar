# REPORT

Журнал создания Duekeep. Записи добавляются по ходу работы, не в конце.

## 2026-08-19 — старт

- Учебное задание: full-stack приложение с нуля (AI-процесс, Docker Compose, OpenAPI, CI).
- Официальные темы (финансы, Trello, рецепты, квизы, заметки, склад) отклонены.
- Выбрана тема: **календарь истечений** (домены, SSL, подписки, договоры, гарантии и т.п.).
- Пожелание по стеку: backend на Go, frontend без предпочтений.
- Создан новый репозиторий `expiry-calendar`, записан черновик функционала в `FUNCTIONAL.md`.
- Код и `ARCHITECTURE.md` сознательно не создавались: сначала уточнение объёма.

### Что сработало

- Тема хорошо ложится на чеклист задания (CRUD, фильтры, дашборд, CSV, роли, журнал, live-уведомления) и при этом полезнее типовых учебных сюжетов.

### Открыто

- Закрыто 2026-08-21: приняты дефолты.

## 2026-08-21 — функционал утверждён, архитектура

- Все 8 дефолтов из `FUNCTIONAL.md` приняты.
- Типы записей сделаны гибкими: справочник `item_kinds` + колонка `items.attrs JSONB`. Общие поля (`expires_at`, цена, статус) остаются колонками — иначе ломаются фильтры и дашборд.
- Написан `ARCHITECTURE.md` до кода: стек, схема БД, API, SSE, seed, фазы коммитов.
- `FUNCTIONAL.md` приведён в соответствие (kinds вместо enum, закрыты вопросы).

### Почему не «вся запись в JSONB»

- Пагинация, индексы по дате и агрегации дашборда проще и надёжнее на колонках.
- JSONB оставляем только для специфики типа (VIN, регистратор и т.п.).
- Новый тип = строка в справочнике, без миграции.

### Следующий шаг

- Фаза 0 плана: каркас compose / backend `/healthz` / пустой фронт / CI. Без расширения требований.

## 2026-08-21 — правки архитектуры

- Версия Go: **1.25** (вместо размытого 1.22+).
- Границы пакетов сменены на классику `handler` / `service` / `repository` (+ `model`).
- Причина смены: учебная сдача и тесты. Доменная нарезка экономит прокладки на маленьком CRUD, но здесь много сценариев (тикер, renew, CSV, audit) и проверяющему проще читать слои. Service не должен быть пустым прокси — только инварианты и оркестрация.

## 2026-08-21 — сессии, типы, PWA

- Seed-типы: убраны `ssl` и `warranty`, добавлен `rent` (Аренда). `subscription` уже был — это «Подписки».
- Auth сначала записали как cookie-сессии. Это оказалось рано: продукт пойдёт в многопользовательский режим.
- Клиент обязан быть PWA (manifest, SW, установка) и уметь Web Push (VAPID) при закрытой вкладке. In-app колокольчик + SSE остаются.

## 2026-08-21 — JWT + refresh

- Откат решения «только сессия». Нужны JWT access (15 мин) и refresh (14 дней, таблица `refresh_tokens`, ротация, revoke family при reuse).
- Refresh отдаём и в JSON (будущие клиенты), и в HttpOnly cookie (браузер/PWA, не класть long-lived в `localStorage`).
- Access в памяти + `Authorization: Bearer`. Для SSE — query `access_token` (ограничение EventSource).
- v1 данные ещё общие; в claims сразу `sub` и `role`, позже `org_id` без смены login/refresh.

## 2026-08-21 — спринты вместо фаз

- План разработки перенесён из «фаз» в `docs/` по образцу `my-chat/docs`.
- Шесть спринтов: каркас → auth/справочники → записи → realtime/обзор → UI/CSV/PWA → сдача.
- У каждого: `sprint-N-plan.md`, `api-sprint-N.md`, `sprint-N-checklist.md`, `known-limitations-sprint-N.md`.
- В `ARCHITECTURE.md` осталась только сводная таблица и ссылки. Код начинается с чеклиста Sprint 1.

## 2026-08-21 — правила Cursor

- Создана папка `.cursor/rules/` по образцу `my-chat`: `go-idioms.mdc` (Go 1.25), `lint.mdc`.
- Добавлен `sprints.mdc`: агент опирается на `docs/sprint-N-*`, не расширяет scope, API правит вместе с контрактом.

## 2026-08-22 — Taskfile

- Корневой `Taskfile.yml` по образцу my-chat: `tools:install`, `tidy`, `fmt` (gofumpt), `lint` (golangci-lint в Docker + frontend), `lint:local`, `test` (`go test -race`).
- `test:integration`, `local:up`, `local:down` — заглушки с понятной ошибкой, дополним в Sprint 1.
- Правило `.cursor/rules/lint.mdc` снова требует только `task lint`.
- Пока нет `backend/`, цели с `dir: backend` падают на precondition — так и задумано.

## 2026-08-22 — Sprint 1, раздел 1 (Подготовка)

- Заведена структура каталогов из `ARCHITECTURE.md` §5: слои backend + папки frontend.
- Добавлен `.env.example` (без реальных секретов).
- `api-sprint-1.md` не меняли. Хендлер появится в разделе 3 — тогда повторная сверка.

## 2026-08-22 — Sprint 1, раздел 2 (Docker)

- Добавлен `docker-compose.yml`: db + backend + frontend. Backend стартует только после `healthy` у Postgres.
- Заглушки Dockerfile, чтобы `up --build` не ждал Go/Vite. Заменим в разделах 3–4.
- Postgres: `16-alpine` — `postgres:16` не скачался (таймаут registry).
- Прогон: `down -v` + `up --build`. Стек поднялся; `:80` отвечает 200. На `:8080` сидел my-chat — его main-service остановлен.
- Taskfile: рабочие `local:up` / `local:down` / `local:down:clean`.

## 2026-08-22 — deploy/ как в my-chat

- `deploy/local` — демо-пароли в compose, для учёбы и сдачи.
- `deploy/test` — Postgres на 55432, tmpfs, под будущие integration-тесты.
- `deploy/prod` — без секретов в git: `env_file` + обязательные `${VAR:?}`, nginx 80/443, certbot, `init-ssl.sh`.
- Корневой `.env.example` — шаблон прода (JWT, Postgres hex-пароль, VAPID, DOMAIN). Живой `.env` и `certbot/conf` в gitignore.
- Корневой `docker-compose.yml` — `include` локального файла, чтобы `docker compose up` из корня не сломался.

## 2026-08-22 — починка task local:down

- После переноса compose задача гасила проект `duekeep`, а жил старый `expiry-calendar` (имя из папки). Вывод был пустой из‑за `silent: true`.
- Теперь `local:up/down` явно `-p duekeep`, плюс снимают leftover `expiry-calendar`. У корневого compose тоже `name: duekeep`. Silent на этих целях выключен.

## 2026-08-22 — порт Postgres на локали

- В `deploy/local` опубликован `15432:5432` (хостовый 5432 занят). Прод без внешнего порта.

## 2026-08-22 — документация инфра + Sprint 1 §3 Backend

- Зафиксированы в `ARCHITECTURE.md` §4/§5/§15, `sprint-1-plan`, `known-limitations`, `FUNCTIONAL.md`, `deploy/README.md`: `deploy/`, проект `duekeep`, порт БД 15432, секреты только в `.env` на проде.
- Backend: chi + pgx + goose, слои handler→service→repository, `GET /healthz` 200/503 по контракту.
- Образ `golang:1.25-alpine`. slog JSON, DSN в логе без пароля.
- Проверка: `task local:up`, `curl localhost:8080/healthz` → `{"status":"ok"}`. `go test` handler ok.

## 2026-08-22 — Sprint 1 §4 Frontend

- Vite 6 + React 18 + TypeScript + Tailwind 4. Заглушка `HomePage`: Duekeep / «Скоро».
- Образ: multi-stage `node:22` → `nginx:1.27-alpine`. Статика из `dist`, SPA fallback.
- nginx проксирует `/api/` и `/healthz` на `backend:8080` (браузер не ходит на `:8080`).
- `npm run lint`, `typecheck`, `build` зелёные.
- Smoke после `task local:up`: `localhost/` 200 (Duekeep), `localhost/healthz` → ok, `/api/v1/nope` уходит в backend (404 chi).

## 2026-08-22 — Sprint 1 §5 CI

- Добавлен `.github/workflows/ci.yml`: Lint (golangci-lint v2.12.2 в `backend/`), Unit Tests (`go test -race`), Frontend (`npm ci` + lint + typecheck).
- Триггеры: любой push/PR и `workflow_dispatch`. Integration-джобы нет — `test:integration` ещё заглушка.
- Попутно починил precondition в Taskfile: при `dir: backend` проверяем `go.mod`, а не `backend/go.mod`.
- Локально те же команды зелёные. Зелёный Actions — после пуша.

## 2026-08-22 — Sprint 1 §6 DoD

- Чистый прогон из корня: `docker compose down -v && docker compose up --build -d --wait`.
- Три сервиса Up, db healthy, новый том. Goose накатил `001_init.sql` (version 1). DSN в логе с `***`.
- `localhost:8080/healthz` и `localhost/healthz` → 200 `{"status":"ok"}`. `localhost/` → заглушка Duekeep.
- `task test` зелёный. Limitations уже в `docs/known-limitations-sprint-1.md`.
- Sprint 1 закрыт. Actions на GitHub — после пуша.

## 2026-08-22 — живой Swagger

- Обязательное условие сдачи: OpenAPI + Swagger UI, не откладывая на Sprint 6 целиком.
- `backend/openapi.yaml` встроен в бинарь (`duekeep.OpenAPISpec`). UI — `swgui/v5emb` на `GET /docs` (редирект на `/docs/`), спека — `GET /openapi.yaml`.
- Без кодогенерации: спека пишется руками, сейчас только `/healthz` и docs.
- nginx фронта и прод: proxy `/docs` и `/openapi.yaml` на backend. Vite dev — те же proxy.
- Тесты: `TestOpenAPISpec`, `TestDocsRedirect`, `TestDocsUI`.
