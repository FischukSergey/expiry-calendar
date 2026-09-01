# Sprint 8 Checklist

Источник: [`sprint-8-plan.md`](sprint-8-plan.md). Код и workflow — после закрытия Sprint 7.

## 1) Bootstrap VPS

- [x] Docker Engine + Compose v2, git на `159.194.252.6`.
  Примечание: Docker 29.7.2, Compose v5.5.0, git 2.53. На 2 ГБ RAM добавлен `/swapfile` 2G.
- [x] Клон репозитория в фиксированный каталог, ветка `main`.
  Примечание: `/opt/duekeep`, `origin` = GitHub, на момент bootstrap SHA `85f2a6c` (Sprint 7). Старый клон `/root/expiry-calendar` не трогали.
- [x] Корневой `.env` (`chmod 600`), `DOMAIN=duekeep.ru`, seed выключен.
- [x] `deploy/prod/nginx/conf.d/duekeep.conf`: `server_name` и пути LE = `duekeep.ru`.
- [x] `init-ssl.sh --staging`, затем боевой сертификат.
  Примечание: staging выпущен и удалён; боевой LE `CN=duekeep.ru`, до 2026-11-30.
- [x] Первый `up -d --build`: `https://duekeep.ru/healthz` → 200.
  Примечание: `{"status":"ok"}`, curl HTTP 200.

## 2) Скрипт деплоя

- [x] `deploy/prod/deploy.sh` принимает SHA или ref, по умолчанию `origin/main`.
- [x] `git fetch` + checkout, затем `docker compose -f deploy/prod/docker-compose.prod.yml --env-file .env up -d --build`.
- [x] Не трогает `.env` и тома Postgres.
- [x] После подъёма: HTTP 200 на `https://duekeep.ru/healthz` (таймаут задан).
  Примечание: `HEALTH_TIMEOUT=600`. Пока скрипт не в `main`, на VPS лежит копия из рабочей ветки.
- [ ] Повторный запуск на том же SHA не ломает стек.
  Примечание: не гонять `deploy.sh origin/main` до мержа Sprint 8 — `checkout` откатит nginx на `example.com`.

## 3) GitHub Actions

- [x] Deploy только после зелёных lint / test / build / frontend **того же SHA**.
  Примечание: job `Deploy` в `ci.yml`, `needs` четырёх job, общий `CHECKOUT_REF`.
- [x] Триггеры: `push` в `main`, `workflow_dispatch`. PR не деплоит.
- [x] Тег `v1.0.0` не запускает прод-деплой.
  Примечание: `if` отсекает тег и PR; скрипт тоже отвергает точный тег `v1.0.0`.
- [ ] Secrets: хост, пользователь, приватный ключ (и порт при необходимости).
  Примечание: ключ `~/.ssh/duekeep_github_actions` создан локально. В GitHub Secrets ещё не заведены (`gh` нет).
- [x] Отдельный deploy-ключ в `authorized_keys` на VPS, не личный `beget_duekeep`.
  Примечание: `restrict,command=/opt/duekeep/deploy/prod/ssh-deploy.sh`, comment `duekeep-github-actions`.
- [x] Проверка host key (fingerprint), не отключение StrictHostKeyChecking.
  Примечание: `deploy/prod/known_hosts`, ED25519 `SHA256:QvNSNgx3ji3Lug0TwOEwUd+XT8gz0TR7EqUpqZ4w++M`.

## 4) Документы и откат

- [x] `deploy/README.md`: bootstrap, что в Actions / что в `.env`, команда отката.
- [x] README: прод обновляется с `main` (ссылка на deploy).
- [ ] Откат проверен: `deploy.sh <предыдущий_sha>`, healthz 200, данные на месте.
  Примечание: после первого деплоя с `main` (уже со скриптом).

## 5) DoD

- [x] Контракт [`api-sprint-8.md`](api-sprint-8.md) (без новых ручек).
- [x] [`known-limitations-sprint-8.md`](known-limitations-sprint-8.md) заполнен.
- [ ] Демо из плана §7 пройдено.
  Примечание: healthz живой; зелёный Deploy в Actions — после секретов и push в `main`.
- [x] `task lint` и `task test` зелёные (репозиторий не сломан скриптом).
