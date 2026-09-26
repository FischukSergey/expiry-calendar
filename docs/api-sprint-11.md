# API контракты Sprint 11

Источник: [`sprint-11-plan.md`](sprint-11-plan.md). База — сумма [`api-sprint-1.md`](api-sprint-1.md)…[`api-sprint-10.md`](api-sprint-10.md). Новых путей нет. Меняются права kinds, код login и валидация push.

## 1) Решения

- Регистрация по-прежнему создаёт пользователя с ролью `admin` и правом писать **свои** записи.
- Справочник `item_kinds` общий. Писать его может только роль `administrator`. Эту роль регистрация не выдаёт. На проде её нет, пока email не указан в `ADMINISTRATOR_EMAIL`.
- `GET /me` по-прежнему отдаёт `id`, `email`, `role`. Новое значение `role`: `administrator`.
- Статус записи для monthly/yearly считается от самого раннего неоплаченного вхождения, включая дату раньше сегодня. Форма ответа item не меняется: те же `status` и `next_open_at`.
- `owner_id` в JSON нет.
- Демо-seed в API не отражён: учёток и записей при старте нет.

## 2) Kinds

`GET /api/v1/kinds` — как раньше, любой auth.

`POST /api/v1/kinds`, `PATCH /api/v1/kinds/{id}`, `DELETE /api/v1/kinds/{id}`:

- нет access → `401`;
- роль `viewer` или `admin` → `403`;
- роль `administrator` → прежние `201` / `200` / `204` и прежние тела.

Чужой предметный id (items, categories, payments) по-прежнему `404`, не `403`.

## 3) Login

`POST /api/v1/auth/login`.

После 8 неудачных попыток на пару IP + email в окне 15 минут:

```json
{"error":{"code":"rate_limited","message":"rate limited"}}
```

Ответ `429`. Успешный login сбрасывает счётчик этой пары. `POST /auth/refresh` лимитом не покрыт. Счётчик в памяти процесса.

## 4) Push

`POST /api/v1/push/subscribe`.

`endpoint` обязан быть `https` и указывать на публичный адрес. Loopback, частные сети, link-local, multicast и пустой хост:

```json
{"error":{"code":"validation_error","message":"invalid subscription","details":{"endpoint":"url"}}}
```

Ответ `422`. Та же проверка повторяется перед отправкой и наружу не торчит отдельным кодом.

## 5) Без изменений контракта

- `POST /api/v1/items/bulk` — тот же body `{ ids, category_id?, status? }`. Появляется только UI.
- Платёж вхождения, календарь, `occurrence_status` — Sprint 10.
- `started_at` позже `expires_at` сервер уже отвечает `422` (`details.started_at`). Клиент начинает проверять то же до запроса.
