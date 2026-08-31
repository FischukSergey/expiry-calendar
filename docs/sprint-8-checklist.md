# Sprint 8 Checklist

Источник: [`sprint-8-plan.md`](sprint-8-plan.md). Код и workflow — после закрытия Sprint 7.

## 1) Bootstrap VPS

- [ ] Docker Engine + Compose v2, git на `159.194.252.6`.
- [ ] Клон репозитория в фиксированный каталог, ветка `main`.
- [ ] Корневой `.env` (`chmod 600`), `DOMAIN=duekeep.ru`, seed выключен.
- [ ] `deploy/prod/nginx/conf.d/duekeep.conf`: `server_name` и пути LE = `duekeep.ru`.
- [ ] `init-ssl.sh --staging`, затем боевой сертификат.
- [ ] Первый `up -d --build`: `https://duekeep.ru/healthz` → 200.

## 2) Скрипт деплоя

- [ ] `deploy/prod/deploy.sh` принимает SHA или ref, по умолчанию `origin/main`.
- [ ] `git fetch` + checkout, затем `docker compose -f deploy/prod/docker-compose.prod.yml --env-file .env up -d --build`.
- [ ] Не трогает `.env` и тома Postgres.
- [ ] После подъёма: HTTP 200 на `https://duekeep.ru/healthz` (таймаут задан).
- [ ] Повторный запуск на том же SHA не ломает стек.

## 3) GitHub Actions

- [ ] Deploy только после зелёных lint / test / build / frontend **того же SHA**.
- [ ] Триггеры: `push` в `main`, `workflow_dispatch`. PR не деплоит.
- [ ] Тег `v1.0.0` не запускает прод-деплой.
- [ ] Secrets: хост, пользователь, приватный ключ (и порт при необходимости).
- [ ] Отдельный deploy-ключ в `authorized_keys` на VPS, не личный `beget_duekeep`.
- [ ] Проверка host key (fingerprint), не отключение StrictHostKeyChecking.

## 4) Документы и откат

- [ ] `deploy/README.md`: bootstrap, что в Actions / что в `.env`, команда отката.
- [ ] README: прод обновляется с `main` (ссылка на deploy).
- [ ] Откат проверен: `deploy.sh <предыдущий_sha>`, healthz 200, данные на месте.

## 5) DoD

- [ ] Контракт [`api-sprint-8.md`](api-sprint-8.md) (без новых ручек).
- [ ] [`known-limitations-sprint-8.md`](known-limitations-sprint-8.md) заполнен.
- [ ] Демо из плана §7 пройдено.
- [ ] `task lint` и `task test` зелёные (репозиторий не сломан скриптом).
