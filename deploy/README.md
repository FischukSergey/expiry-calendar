# Deploy Duekeep

Как в my-chat: локальный стек, тестовая БД и prod на VPS разнесены по каталогам.

## Хост прода (Beget)

| | |
|---|---|
| Домен | `duekeep.ru` (без `www`) |
| IPv4 VPS | `159.194.252.6` |
| Каталог | `/opt/duekeep`, ветка/SHA с `origin` |
| SSH вручную | `ssh duekeep` (ключ `~/.ssh/beget_duekeep`, алиас в `~/.ssh/config`) |
| SSH CI | отдельный ключ `duekeep-github-actions` в `authorized_keys`, не личный |
| DNS | A `@` → `159.194.252.6`, NS Beget. `www` / autoconfig в зону не нужны |
| SSL | свой certbot на VPS (`init-ssl.sh`), не бесплатный LE в панели Beget |

В `.env` на сервере: `DOMAIN=duekeep.ru`. То же имя — в [`prod/nginx/conf.d/duekeep.conf`](prod/nginx/conf.d/duekeep.conf) (`server_name` и пути Let's Encrypt).

| Каталог | Назначение | Секреты |
|---|---|---|
| [`local/`](local/docker-compose.local.yml) | разработка и сдача (`docker compose up` из корня) | демо, зашиты в compose |
| [`test/`](test/docker-compose.test.yml) | Postgres для будущих integration-тестов | демо |
| [`prod/`](prod/docker-compose.prod.yml) | VPS | **только** корневой `.env`, не в git |

## Секреты: что где

| Где | Что |
|---|---|
| `.env` на VPS (`chmod 600`) | Postgres, JWT, VAPID, `DOMAIN`, `LETSENCRYPT_EMAIL` |
| GitHub Actions Secrets | только SSH: `DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_SSH_KEY`, при необходимости `DEPLOY_PORT` |

JWT, пароль БД и VAPID в Actions не кладём. Утечка deploy-ключа = SSH на сервер; ключ снимают из `authorized_keys` и ротируют секрет.

В git не должно быть: `.env`, живых JWT/VAPID/паролей, каталога `deploy/prod/certbot/conf/` (сертификаты).

Локальный стек `.env` не читает — там намеренно слабые демо-значения. `SEED=true`.
Прод: `SEED=false` в compose — демо-пользователи и 50+ items не создаются.

Локальная Postgres с хоста: `localhost:15432`, пользователь/пароль/БД `duekeep`.
Внутри compose backend ходит на `db:5432` (имя сервиса).
На проде порт наружу **не** публикуем.

## Bootstrap VPS (один раз)

Когда `dig +short duekeep.ru A` даёт `159.194.252.6`:

1. Docker Engine + Compose v2, git, curl, openssl.
2. На 1–2 ГБ RAM без swap сборка образов часто падает с 137 — файл подкачки 2G (`/swapfile`) обязателен на таком тарифе.
3. Клон в `/opt/duekeep`, remote `origin`, стартовая ветка `main`.
4. `cp .env.example .env`, заполнить секреты, `chmod 600 .env`. `DOMAIN=duekeep.ru`. Seed в compose уже выключен.
5. `duekeep.conf`: `server_name` и пути LE = `duekeep.ru`.
6. SSL, затем стек:

```bash
# в .env: DOMAIN=duekeep.ru и LETSENCRYPT_EMAIL=...
cd /opt/duekeep/deploy/prod
bash init-ssl.sh --staging   # проверка
bash init-ssl.sh             # боевой сертификат
```

7. Проба: `https://duekeep.ru/healthz` → 200.

Первый сертификат только с сервера, не из Actions.

## CD

После зелёного lint / test / build / frontend **того же SHA** job Deploy в [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) по SSH запускает [`prod/deploy.sh`](prod/deploy.sh).

Триггеры: `push` в `main`, `workflow_dispatch` (можно указать SHA). PR и тег `v1.0.0` прод не меняют.

Скрипт: `git fetch` + `checkout --detach`, затем

```bash
docker compose -f deploy/prod/docker-compose.prod.yml --env-file .env up -d --build
```

`.env` и том Postgres не трогает. После подъёма ждёт HTTP 200 на `https://duekeep.ru/healthz` (таймаут 600 с). Повтор на том же SHA безопасен.

Вручную с ноутбука:

```bash
ssh duekeep '/opt/duekeep/deploy/prod/deploy.sh'
ssh duekeep '/opt/duekeep/deploy/prod/deploy.sh <sha>'
```

На самой VPS: `bash /opt/duekeep/deploy/prod/deploy.sh` или `task prod:deploy`.

### Ключ Actions

Отдельная пара ed25519, комментарий `duekeep-github-actions`. В `authorized_keys` — forced command, не интерактивный shell:

```
restrict,command="/opt/duekeep/deploy/prod/ssh-deploy.sh" ssh-ed25519 AAAA... duekeep-github-actions
```

Клиент (Actions) передаёт только SHA или ref. Host key — [`prod/known_hosts`](prod/known_hosts), в workflow `StrictHostKeyChecking=yes`.

Секреты репозитория:

| Secret | Пример |
|---|---|
| `DEPLOY_HOST` | `159.194.252.6` |
| `DEPLOY_USER` | `root` |
| `DEPLOY_SSH_KEY` | приватный ключ целиком (включая `BEGIN`/`END`) |
| `DEPLOY_PORT` | `22`, если не 22 |

Fingerprint ED25519 хоста: `SHA256:QvNSNgx3ji3Lug0TwOEwUd+XT8gz0TR7EqUpqZ4w++M`.

## Откат

Тот же скрипт и те же `.env` / тома:

```bash
ssh duekeep '/opt/duekeep/deploy/prod/deploy.sh <предыдущий_sha>'
```

или `workflow_dispatch` с этим SHA. `v1.0.0` скрипт отклоняет (общий каталог сдачи). Автоотката нет: красный healthz — выкладка руками.
