# REPORT

Журнал создания Duekeep. Записи добавляются по ходу работы, не в конце.

## 2026-08-31 — Sprint 7 §2, auth

- Register создаёт admin; login/refresh/me без `org_id` в claims и `/me`.
- Patch/delete/renew/bulk items и patch/delete categories — только свой `owner_id`, иначе 404.

## 2026-08-31 — Sprint 7 §1, owner_id

- Миграция `011_owner_id.sql`: колонка на categories/items/audit_log/notifications, backfill seed-admin, NOT NULL + FK + индексы. Org-таблиц нет.
- Seed и INSERT в repository пишут `owner_id`, иначе NOT NULL ломает compose. Выборка по владельцу — следующие пункты.

## 2026-08-31 — Sprint 8 (документы, без кода)

- CD на Beget: после CI на `main` — SSH, `deploy.sh`, `compose --build`, `https://duekeep.ru/healthz`.
- Не входит: registry, zero-downtime, секреты приложения в Actions, деплой `v1.0.0`.
- Добавлены `docs/sprint-8-*.md`, строки в индексе и ARCHITECTURE.

## 2026-08-31 — Sprint 7, смена цели

- Прод: register → admin, видит только свои строки (`owner_id` = `sub`).
- Не делаем: viewer, инвайты, `org_id` / org / шаринг.
- Обновлены plan, checklist, api, limitations спринта 7; указатели в ARCHITECTURE, FUNCTIONAL, docs/README, README.

## 2026-08-30 — хост прода

- Документированы домен `duekeep.ru` и IPv4 VPS Beget `159.194.252.6` (`deploy/README.md`, `.env.example`, README).

## 2026-08-30 — дашборд: суммы по месяцам

- Столбцы «Сроки оплаты» заменены на «Суммы оплаты по месяцам».
- API: у `expirations_by_month` добавлен `amounts` (сумма `cost_amount` по валютам, без конвертации). `count` сохранён.

## 2026-08-29 — Sprint 6 старт

- Спринты 1–5 закрыты. Дыры плана (reuse refresh, CSV dry_run, calendar, push 410) уже покрыты тестами; CI lint+test+build+frontend на месте; `/docs` живой.
- Не хватает объёма seed (4 items вместо 50+), renewals/audit/unread, сверки OpenAPI cookie refresh, README под сдачу.
- Дальше: каталог seed, спека, README, прогон на чистом томе.

## 2026-08-29 — Sprint 6, seed и OpenAPI

- Каталог: 52 items (даты от Clock.Today), ≥5 expired, ≥8 expiring, 1 cancelled, 1 archived. 22 renewals, 24 audit, unread на expired/expiring. Повторный seed не плодит строки; даты items/renewals обновляет.
- CheckCatalog требует объём FUNCTIONAL. OpenAPI 1.0.0: cookie duekeep_refresh на refresh/logout, body важнее cookie. README под сдачу.

## 2026-08-30 — правки формулировок формы

- Обязательные поля на «Новая запись» со звёздочкой (название, тип, срок оплаты).
- Тип записи и раздел разделены подсказками; при выборе типа подставляется раздел. В списке колонка раздела убрана.
- «Начало» → «Начало периода». «Истекает/истекло» в UI → «срок оплаты» / «скоро срок» / «просрочено».

## 2026-08-29 — фикс логина

- TextInput/Select/TextArea не пробрасывали ref (React 18). register не видел поля → Zod «expected string, received undefined» на заполненной форме. Теперь forwardRef.

## 2026-08-29 — Sprint 6 follow-up (демо)

- Смена `expires_at` в админке пишет notification и SSE: иначе тикер не видит переход (статус уже пересчитан при PATCH). `TestPatchExpiresNotifies`.
- VAPID в `deploy/local` зафиксирован — подписки не сбрасываются после рестарта backend.

## 2026-08-29 — Sprint 6 закрыт

- Чистый том: `docker compose down -v && docker compose up --build`. db/backend/frontend healthy.
- Живые цифры: 2 users, 9 kinds, 13 categories, 52 items (6 expired, 12 expiring), 22 renewals, 24 audit, 18 unread. Повторный up — те же счётчики.
- `/healthz` ok, `/docs` Swagger, UI `:80`, viewer 403 на запись. Calendar текущего месяца не пустой, CSV export отдаёт фильтр.
- `task lint` / `task test` зелёные (83 теста). Limitations v1 без изменений. Код спринта ещё не закоммичен.

## 2026-08-29 — Sprint 5 закрыт

- §4: адаптив (сайдбар / табы, карточки списка, календарь, safe-area), loading/error/empty + «Повторить».
- Контракт CSV сверен с роутером и OpenAPI. Limitations: нет офлайн-CRUD, seed 50+ и полная сверка спеки — Sprint 6.
- DoD: login/dashboard/список/карточка/календарь/CSV; viewer без кнопок записи; PWA-артефакты на localhost. `task lint` / `task test` зелёные.

## 2026-08-29 — Sprint 5, раздел 3 (Realtime и PWA)

- После логина: EventSource на `/events?access_token=`. Событие `notification` инвалидирует уведомления, список, дашборд, календарь. Истёкший access — refresh и новое соединение.
- Пуши: разрешение Notification, VAPID, subscribe; выход снимает подписку. SW показывает системное уведомление, клик открывает карточку.
- PWA: `vite-plugin-pwa` injectManifest, manifest Duekeep standalone, иконки 192/512, Workbox network-first для HTML/API (без SSE), офлайн-заглушка «нет сети».
- `beforeinstallprompt` — баннер «Установить» и кнопка в профиле. `task lint` / сборка зелёные.

## 2026-08-29 — Sprint 5, раздел 2 (Экраны)

- SPA: React Router, TanStack Query, react-hook-form + zod, Recharts. Тёмная тема как у заглушки Sprint 1.
- Access в памяти; refresh/logout с `credentials: 'include'`; на 401 — один refresh и повтор. Холодный старт — refresh по cookie.
- Экраны: вход/регистрация, дашборд, список (фильтры в URL + экспорт), форма + attrs, карточка/renew/история, календарь, категории, уведомления, аудит, импорт CSV, профиль.
- Viewer не видит кнопок записи; импорт/аудит/форма закрыты маршрутом.
- SSE, пуши и PWA — раздел 3. `npm run lint` / `typecheck` зелёные.

## 2026-08-29 — Sprint 5, раздел 1 (CSV API)

- `GET /items/export`: тот же фильтр, что список; без пагинации, max 10_000; колонки + `attrs.*` из schema видов в выгрузке. Viewer можно.
- `POST /items/import`: multipart `file` + JSON `mapping` (поле → колонка, включая `attrs.*`). `dry_run=true` — превью без записи; запись — одна транзакция и audit `import`. Ошибки строк — 422, ничего не пишем. Потолок 5_000.
- Хендлер тонкий; маппинг и coerce attrs — unit в `service`. `task test` / `task lint` зелёные.

## 2026-08-29 — Sprint 4 закрыт

- DoD: Tick меняет статус и пишет notification; SSE видит event; dashboard отдаёт series (6 месяцев, soonest).
- HTTP: `TestTickerIntegrationStatusAndNotification`, `TestDashboardViewerTwoCurrencies` (RUB и USD не сливаются).
- Контракт сверен с роутером и OpenAPI. Limitations: UI в Sprint 5, SSE один процесс, integration без Postgres.
- `task test` / `task lint` зелёные. Код Sprint 4 ещё не закоммичен.

## 2026-08-29 — Sprint 4, раздел 4 (Обзор)

- `GET /dashboard`: counts, upcoming_cost по валютам без конвертации, 6 месяцев истечений, cost_by_kind, soonest (10).
- `GET /calendar?year=&month=`: дни только с записями. cancelled/archived не входят.
- Один `ListOpen`, агрегаты в service. `task test` / `task lint` зелёные.

## 2026-08-29 — Sprint 4, раздел 3 (Web Push)

- Миграция `010_push_subscriptions.sql`: unique `endpoint`, индекс по `user_id`.
- API: `GET /push/vapid-public`, `POST`/`DELETE /push/subscribe` (auth, viewer). Upsert по endpoint; unsubscribe идемпотентен.
- Тикер через `service.Fanout`: после INSERT — SSE и Web Push всем подпискам (данные общие). Payload как у SSE.
- `410 Gone` от push-сервиса удаляет строку. `webpush-go`; VAPID из env, иначе генерация на процесс.
- `task test` / `task lint` зелёные.

## 2026-08-28 — долг после v1: конфиг

- Тикер 60 с оставить до сдачи; после v1 сменить (домен — день, не минута).
- Не-секреты (HTTP, TTL, интервал тикера, SSE ping и т.п.) — файл конфига при инициализации, не только env.
- В `.env` только чувствительное: пароли, JWT, VAPID. Зафиксировано в `known-limitations-sprint-4.md` и sprint-6.

## 2026-08-28 — Sprint 4, раздел 2 (SSE)

- `GET /api/v1/events`: Bearer или `?access_token=`. `text/event-stream`, сразу `ping`, дальше каждые 15 с.
- Hub в памяти, mutex; полный буфер клиента не блокирует тикер.
- Тикер после успешного INSERT шлёт `event: notification`. Повтор в тот же день — без события.
- nginx фронта: `/api/v1/events` без буфера (как в prod). `task test` / `task lint` зелёные.

## 2026-08-28 — Sprint 4, раздел 1 (Статусы и уведомления)

- Тикер в `cmd/server`: сразу `Tick`, затем каждые 60 с. Тот же `StatusAtWrite`, что при записи. `cancelled` / `archived` не трогает.
- Миграция `009_notifications.sql`: unique `(item_id, to_status, день UTC)`. Повторный tick в тот же день не плодит строки.
- API: `GET /notifications` (`unread`, пагинация как у items), `POST /{id}/read`, `POST /read-all`. Viewer и admin.
- Тесты зовут `Tick` явно, без ожидания минуты. `task test` / `task lint` зелёные.

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

## 2026-08-22 — CI build

- В `.github/workflows/ci.yml` джоба Build: `go build ./cmd/server`.
- Frontend: к lint/typecheck добавлен `npm run build`.
- `docker compose build` в CI по-прежнему нет — ловим Dockerfile локально.

## 2026-08-25 — Sprint 2, раздел 1 (Данные)

- Миграции goose 002–005: `users`, `refresh_tokens`, `item_kinds`, `categories` по схеме `ARCHITECTURE.md` §6.2.
- Идемпотентный seed в `internal/seed`: 2 пользователя (фиксированные UUID из api-sprint-2), 9 kinds без ssl/warranty, 13 категорий (глубина 2).
- `cmd/server`: goose up → seed → listen. Конфликт по email/slug/id — `ON CONFLICT DO NOTHING`.
- Проверка: `task test` / `task lint` зелёные. Compose: goose version 5, повторный restart не плодит строки.
- Демо-пароли записаны в README, не как прод-секрет.

## 2026-08-25 — Sprint 2, раздел 2 (Auth)

- JWT HS256 access 15 мин, opaque refresh 14 дней. Claims: sub, role, iss=duekeep, iat, exp.
- Register всегда viewer. Refresh: body важнее cookie. Reuse отозванного → revoke family; неизвестный токен семьи не трогает.
- Bearer только на /me и logout-all. Cookie Path=/api/v1/auth, без Secure на локали.
- JWT_SECRET обязателен, порог «≥ 32» не вводили (local compose короче).

## 2026-08-25 — Sprint 2, раздел 3 (Справочники)

- CRUD kinds/categories: GET под Bearer, мутации — admin (`RequireAdmin`). Viewer → 403.
- `attr_schema`: массив `{key,label,type,required}`, type string|number|boolean, уникальный key.
- Дерево категорий: глубина ≤ 3 (корень = 1), цикл и высота поддерева при move; DELETE с детьми → 409.
- `CountItems` = 0 до таблицы items. OpenAPI дополнен путями kinds/categories.

## 2026-08-25 — Sprint 2, разделы 4–5 (тесты и DoD)

- Unit: hash refresh (SHA-256 hex), claims access (sub/role/iss/iat/exp), глубина дерева и цикл.
- HTTP-сценарии на реальном service + память: login→refresh→logout, reuse family, viewer 403 на POST kind, admin создаёт kind.
- Контракт `api-sprint-2.md` сверен с хендлерами. Limitations: CountItems=0, нет UI логина, нет Postgres в `task test:integration`.
- `task lint` / `task test` зелёные. Compose после rebuild: login admin → /me /kinds(9) /categories → refresh; старый refresh 401; viewer POST kind 403.
- Sprint 2 закрыт по чеклисту.

## 2026-08-26 — Sprint 3, разделы 4–5 (тесты и DoD)

- Unit: attrs, статус при записи, глубина категорий — уже были и остаются зелёными.
- HTTP: `TestItemsCRUDRenewFilterPage` (CRUD, renew пишет историю, q+page, tag), `TestViewerForbiddenItemMutations`.
- Контракт `api-sprint-3.md` сверен с роутером и OpenAPI. Limitations: нет тикера/UI, audit без фильтров, тесты без Postgres.
- `task test` / `task lint` зелёные. Sprint 3 закрыт по чеклисту.

## 2026-08-26 — Sprint 3, раздел 3 (Аудит)

- Снимок `itemAuditSnap`: только id, title, kind_id, category_id, status, expires_at, cost_amount, attrs.
- Create/update/delete/renew/bulk пишут audit в той же транзакции. Тест `TestMutationsWriteAudit` проверяет все action и отсутствие url/account_hint/паролей.

## 2026-08-26 — Sprint 3, раздел 2 (API)

- CRUD `/api/v1/items`, фильтры + CTE потомков категории, пагинация, renew, bulk, GET `/audit`.
- Валидация `attrs` против `attr_schema`; статус при записи от `clock.Today` (UTC), с клиента только cancelled/archived.
- Viewer 403 на мутации и audit; OpenAPI дополнен путями items/audit.
- Мутации пишут audit (before/after без секретов) — задел на раздел 3.

## 2026-08-26 — Sprint 3, раздел 1 (Данные)

- Миграции goose 006–008: `items` (гибрид колонки + `attrs` JSONB, индексы и GIN), `renewals`, `audit_log`.
- Seed: 4 записи с фиксированными UUID, даты относительно `clock.Today` (UTC), статус `active`/`expiring`/`expired` при вставке. Повторный restart обновляет даты, не плодит строки.
- `CountItems` для kind/category теперь считает строки в `items` — DELETE занятого справочника даст 409, а не пройдёт мимо.
- Полный каталог 50+ items — Sprint 6. Notifications — Sprint 4.
- `cost_amount` / `renewals.old_cost` / `new_cost` — `INT`, без дробной части.

## 2026-08-25 — Sprint 7 (документы, без кода)

- После v1: хозяин своей org на одном сервере (PWA), инвайт viewer без почты.
- Добавлены `docs/sprint-7-*.md`, строка в индексе спринтов. Код — только после Sprint 6.
- Спринты 1–6 не меняем: v1 остаётся общим контуром + роли.
