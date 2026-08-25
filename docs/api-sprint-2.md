# API контракты Sprint 2

Источник: [`sprint-2-plan.md`](sprint-2-plan.md). `/healthz` без изменений — [`api-sprint-1.md`](api-sprint-1.md).

## 1) Решения

- Access JWT 15 мин, refresh 14 дней, ротация, family revoke.
- Refresh в JSON и в cookie `duekeep_refresh`.
- Непустой `refresh_token` в body важнее cookie (один источник истины в service — сырая строка).
- Регистрация → роль `viewer`. Пароль короче 8 символов → 422.
- Справочники общие на инсталляцию (не per-user).

## 2) Соглашения

- Base: `/api/v1`
- Защищённые ручки: `Authorization: Bearer <access_token>`
- Ошибки — конверт Sprint 1. Коды: `unauthorized` (401), `forbidden` (403), `not_found` (404), `conflict` (409), `validation_error` (422), `internal` (500)

## 3) Auth

### `POST /api/v1/auth/register`

```json
{ "email": "new@duekeep.local", "password": "secret12" }
```

`201` — тот же объект, что login. Роль `viewer`.

### `POST /api/v1/auth/login`

```json
{ "email": "admin@duekeep.local", "password": "..." }
```

`200`:

```json
{
  "access_token": "eyJ...",
  "refresh_token": "opaque...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

Set-Cookie: `duekeep_refresh=...; HttpOnly; Path=/api/v1/auth; SameSite=Lax`.

### `POST /api/v1/auth/refresh`

Body опционален, если есть cookie. Если есть оба — берём body.

```json
{ "refresh_token": "opaque..." }
```

`200` — новая пара, старый refresh недействителен.

`401` — неизвестный, просроченный, revoked или reuse family.

### `POST /api/v1/auth/logout`

Access и/или refresh. Нужен хотя бы один. `204`, в том числе если refresh уже неизвестен или отозван (не revoke family).

### `POST /api/v1/auth/logout-all`

Только access. `204`.

### `GET /api/v1/me`

`200`:

```json
{
  "id": "11111111-1111-1111-1111-111111111111",
  "email": "admin@duekeep.local",
  "role": "admin"
}
```

## 4) Kinds

### `GET /api/v1/kinds`

Auth. `200`: `{ "items": [ { "id", "slug", "name", "color", "attr_schema" } ] }`

`attr_schema`:

```json
[
  { "key": "registrar", "label": "Регистратор", "type": "string", "required": false }
]
```

`type`: `string` | `number` | `boolean`.

### `POST /api/v1/kinds`

Admin. `201` созданный kind.

### `PATCH /api/v1/kinds/{id}`

Admin. Частичное обновление.

### `DELETE /api/v1/kinds/{id}`

Admin. `204` или `409`, если есть items.

## 5) Categories

### `GET /api/v1/categories`

Дерево: элементы с `id`, `parent_id`, `name`, `sort_order`, `children`.

### `POST /api/v1/categories`

Admin: `{ "parent_id": null, "name": "IT", "sort_order": 0 }`

`422`, если глубина стала бы > 3.

### `PATCH /api/v1/categories/{id}`

Admin. Смена родителя с теми же инвариантами.

### `DELETE /api/v1/categories/{id}`

`204` или `409` (дети или items).

## 6) Совместимость

Хендлеры Sprint 3+ не меняют поля auth/kinds/categories без правки этого файла.
