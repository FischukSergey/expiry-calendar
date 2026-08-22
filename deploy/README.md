# Deploy Duekeep

Как в my-chat: локальный стек, тестовая БД и prod на VPS разнесены по каталогам.

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

Когда будет домен и A-запись:

```bash
# в .env: DOMAIN=... и LETSENCRYPT_EMAIL=...
cd deploy/prod
bash init-ssl.sh --staging   # проверка
bash init-ssl.sh             # боевой сертификат
```

Домен в nginx (`nginx/conf.d/duekeep.conf`) должен совпадать с `DOMAIN`.
