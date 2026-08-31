# Sprint 8 — CD на VPS

Источник: [`deploy/README.md`](../deploy/README.md), хост `duekeep.ru` / `159.194.252.6`. Спринты 1–7 не переписываем. Сдача защиты — тег `v1.0.0`, этот workflow её не трогает.

## 1) Цель

После зелёного CI на `main` прод на Beget обновляется **без ручного SSH**: pull нужного SHA, `docker compose … up -d --build`, проверка `https://duekeep.ru/healthz`.

К концу спринта:

- одноразовый bootstrap VPS описан и пройден (Docker, клон, `.env`, SSL);
- скрипт деплоя в репозитории, идемпотентный;
- GitHub Actions деплоит только `main` (и ручной `workflow_dispatch`), не PR и не тег `v1.0.0`;
- секреты приложения остаются в `.env` на сервере, в Actions — только SSH;
- `deploy/README.md` содержит порядок CD и отката.

## 2) Входные условия

- Sprint 7 закрыт: `owner_id`, register → admin, seed на prod выключен.
- DNS: A `duekeep.ru` → `159.194.252.6`.
- На VPS уже можно зайти (`ssh duekeep`).

## 3) Границы

### Входит

- ручной первый подъём: Docker Engine + Compose v2, `git clone`, `.env`, `duekeep.conf` = `duekeep.ru`, `init-ssl.sh`, первый `prod:up`;
- `deploy/prod/deploy.sh`: fetch, checkout SHA, `compose --env-file .env up -d --build`, curl healthz;
- workflow CD в GitHub Actions (после зелёного lint/test/build/frontend на том же SHA);
- deploy-ключ только для Actions (не личный `beget_duekeep`);
- ручной запуск из Actions и с ноутбука (`ssh duekeep` + скрипт);
- документация отката: предыдущий SHA / тег `v2.x`, тот же скрипт.

### Не входит

- Kubernetes, Ansible, отдельный registry (GHCR);
- zero-downtime / blue-green / второй инстанс backend;
- класть JWT, Postgres, VAPID в GitHub Secrets;
- автодеплой с PR и с тега сдачи `v1.0.0`;
- первый выпуск сертификата из Actions (только `init-ssl.sh` с сервера);
- мониторинг, бэкапы БД, автопродление как новый сервис (certbot в compose уже есть);
- новые прикладные API.

## 4) Backlog

### A. Bootstrap (один раз)

- Docker + `docker compose version` v2, git, openssl.
- Клон `main` в фиксированный путь (например `/opt/duekeep`).
- `.env` из example, `chmod 600`, `DOMAIN=duekeep.ru`.
- nginx `server_name` и пути LE = `duekeep.ru`.
- `init-ssl.sh --staging`, затем боевой.
- Первый `up -d --build`, `https://duekeep.ru/healthz` → ok.

### B. Скрипт

- `deploy/prod/deploy.sh [sha|ref]`: по умолчанию `origin/main`.
- Не читает и не перезаписывает `.env`.
- Не вызывает seed (prod compose Sprint 7).
- Ненулевой exit, если healthz не 200 за таймаут.

### C. Actions

- Триггеры: `push` в `main`, `workflow_dispatch` (SHA или ref).
- Deploy job: `needs` зелёных CI-job того же SHA (job в `ci.yml` или reusable workflow). Не `workflow_run` на чужой SHA.
- PR не деплоит. Тег `v1.0.0` не деплоит.
- Secrets: `DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_SSH_KEY`, при необходимости `DEPLOY_PORT`.
- На сервере: `authorized_keys` для этого ключа, команда — только скрипт деплоя.
- `known_hosts` / fingerprint в workflow, не `StrictHostKeyChecking=no`.

### D. Документы

- `deploy/README.md`: bootstrap, секреты Actions vs `.env`, как откатиться.
- README: одна строка, что прод обновляется с `main`.

## 5) Техрешения

- Сборка образов **на VPS** (`--build`), как сейчас в `task prod:up`. Отдельный registry не заводим.
- Источник истины на сервере — git SHA, не «последний pull вслепую» без записи в лог.
- Один инстанс: короткий простой на время rebuild допустим.
- Откат = `deploy.sh <предыдущий_sha>` на том же `.env` и томах Postgres.

## 6) DoD

- push в `main` (после зелёного CI) обновляет `https://duekeep.ru/healthz` и UI;
- `workflow_dispatch` умеет выкатить выбранный SHA;
- PR и `v1.0.0` прод не меняют;
- `.env` и ключи не в git;
- контракт [`api-sprint-8.md`](api-sprint-8.md) (новых ручек нет);
- [`known-limitations-sprint-8.md`](known-limitations-sprint-8.md) заполнен.

## 7) Демо

1. Изменить README или healthz-текст не нужно — достаточно пустого коммита в `main` или `workflow_dispatch`.
2. Actions: CI зелёный → Deploy зелёный.
3. Браузер: `https://duekeep.ru` и `/healthz`.
4. На VPS: `docker compose -f deploy/prod/docker-compose.prod.yml --env-file .env ps` — все сервисы up.

## 8) Риски

- деплой до готовности DNS/SSL → Let's Encrypt и healthz падают;
- Actions с `main` до закрытия Sprint 7 выкатит общий каталог v1 на прод — не включать workflow, пока seed на prod выключен;
- пересборка на 1–2 ГБ RAM → OOM 137;
- общий SSH-ключ в секретах и у людей — компрометация = доступ на сервер;
- `compose up --build` без pin SHA откатится только вручную.

## 9) Артефакты

- `deploy/prod/deploy.sh`;
- job/workflow CD;
- обновлённый `deploy/README.md`;
- docs спринта.
