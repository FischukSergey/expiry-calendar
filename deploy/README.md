# Deploy Duekeep

Как в my-chat: локальный стек, тестовая БД и prod на VPS разнесены по каталогам.

## Хост прода (Beget)

| | |
|---|---|
| Домен | `duekeep.ru` (без `www`) |
| IPv4 VPS | `159.194.252.6` |
| SSH | `ssh duekeep` (ключ `~/.ssh/beget_duekeep`, алиас в `~/.ssh/config`) |
| DNS | A `@` → `159.194.252.6`, NS Beget. `www` / autoconfig в зону не нужны |
| SSL | свой certbot на VPS (`init-ssl.sh`), не бесплатный LE в панели Beget |

В `.env` на сервере: `DOMAIN=duekeep.ru`. То же имя — в [`prod/nginx/conf.d/duekeep.conf`](prod/nginx/conf.d/duekeep.conf) (`server_name` и пути Let's Encrypt).

| Каталог | Назначение | Секреты |
|---|---|---|
| [`local/`](local/docker-compose.local.yml) | разработка и сдача (`docker compose up` из корня) | демо, зашиты в compose |
| [`test/`](test/docker-compose.test.yml) | Postgres для будущих integration-тестов | демо |
| [`prod/`](prod/docker-compose.prod.yml) | VPS | **только** корневой `.env`, не в git |

## Секреты на проде

1. На VPS в корне репозитория: `cp .env.example .env`
2. Заполнить реальные значения (генерация — в комментариях `.env.example`).
3. `chmod 600 .env` и владелец — пользователь деплоя.
4. Запуск только так:

```bash
docker compose -f deploy/prod/docker-compose.prod.yml --env-file .env up -d --build
```

или `task prod:up`.

В git не должно быть: `.env`, живых JWT/VAPID/паролей, каталога `deploy/prod/certbot/conf/` (сертификаты).

Локальный стек `.env` не читает — там намеренно слабые демо-значения.

Локальная Postgres с хоста: `localhost:15432`, пользователь/пароль/БД `duekeep`.  
Внутри compose backend ходит на `db:5432` (имя сервиса).  
На проде порт наружу **не** публикуем.

## Первый SSL на VPS

Когда `dig +short duekeep.ru A` даёт `159.194.252.6`:

```bash
# в .env: DOMAIN=duekeep.ru и LETSENCRYPT_EMAIL=...
cd deploy/prod
bash init-ssl.sh --staging   # проверка
bash init-ssl.sh             # боевой сертификат
```

Домен в nginx (`nginx/conf.d/duekeep.conf`) должен совпадать с `DOMAIN`.

Автодеплой с `main` (скрипт + GitHub Actions) — [Sprint 8](../docs/sprint-8-plan.md). Пока спринт не закрыт, выкладка ручная: `ssh duekeep` и команда compose выше.
