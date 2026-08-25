# API контракты Sprint 7

Источник: [`sprint-7-plan.md`](sprint-7-plan.md). База — сумма [`api-sprint-1.md`](api-sprint-1.md)…[`api-sprint-6.md`](api-sprint-6.md). Login/refresh **не ломаем**: только новые поля и ручки.

## 1) Решения

- Изоляция на одном сервере: текущий `org_id` из access JWT.
- Регистрация без инвайта → хозяин (admin) новой org.
- Регистрация / accept с инвайтом → viewer чужой org.
- Справочник `item_kinds` общий. Категории и записи — per-org.
- Инвайт — opaque token, без email.

## 2) Соглашения

- Base: `/api/v1`
- Защищённые ручки: `Authorization: Bearer <access_token>`
- Коды ошибок без изменений. Чужие id предметных сущностей → `404 not_found`.

## 3) Claims access (аддитивно)

```json
{
  "sub": "<user uuid>",
  "role": "admin",
  "org_id": "<org uuid>",
  "iss": "duekeep",
  "iat": 0,
  "exp": 0
}
```

`role` — роль **в этой** org. Протокол login/refresh тот же.

## 4) Auth (изменения)

### `POST /api/v1/auth/register`

```json
{ "email": "new@duekeep.local", "password": "secret12", "invite_token": null }
```

`invite_token` опционален.

- без токена: `201` пара токенов, пользователь — `admin` новой org;
- с валидным токеном: `201` пара токенов, `viewer` org инвайта, инвайт погашен;
- занятый email → `409 conflict`.

Тело ответа как login (Sprint 2).

### `GET /api/v1/me`

`200`:

```json
{
  "id": "11111111-1111-1111-1111-111111111111",
  "email": "admin@duekeep.local",
  "role": "admin",
  "org_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "org_name": "Демо"
}
```

## 5) Org и инвайты

### `GET /api/v1/org`

Auth. Текущая org: `{ "id", "name", "role" }`.

### `POST /api/v1/org/invites`

Admin текущей org.

`201`:

```json
{
  "token": "opaque…",
  "expires_in": 604800
}
```

Сырой `token` показывается один раз. В БД только hash.

### `POST /api/v1/org/invites/accept`

Auth (access).

```json
{ "token": "opaque…" }
```

`204`. Дальше клиент делает refresh (или сервер отдаёт новую пару — зафиксировать в реализации и здесь). Пользователь становится `viewer`. Уже член → `409`. Токен мёртв → `401` или `422`.

## 6) Предметные ручки

Пути Sprint 3–5 не меняются. Сервер всегда фильтрует по `org_id` из токена.

Нельзя передать чужой `org_id` query/body, чтобы обойти scope.

`GET/POST /kinds` — как Sprint 2 (общий справочник). Кто может писать kinds — [`known-limitations-sprint-7.md`](known-limitations-sprint-7.md).

## 7) Совместимость

- Поля ответов Sprint 2–5 не удалять.
- `/me` и JWT только расширяются.
- Хендлеры после Sprint 7 не отдают данные без `org_id` в запросе к БД.
