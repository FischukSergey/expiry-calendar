# API контракты Sprint 7

Источник: [`sprint-7-plan.md`](sprint-7-plan.md). База — сумма [`api-sprint-1.md`](api-sprint-1.md)…[`api-sprint-6.md`](api-sprint-6.md). Login/refresh **не ломаем**. Новых полей JWT нет. Меняется роль после register и scope данных.

## 1) Решения

- Изоляция на одном сервере: `owner_id` = `sub` из access. Не `org_id`.
- Регистрация → `admin`, пустой свой каталог (дефолтные категории).
- Нет инвайтов, нет `viewer` как роли шаринга.
- Справочник `item_kinds` общий. Категории и записи — per-user.

## 2) Соглашения

- Base: `/api/v1`
- Защищённые ручки: `Authorization: Bearer <access_token>`
- Коды ошибок без изменений. Чужие id предметных сущностей → `404 not_found`.

## 3) Claims access

Как Sprint 2, без `org_id`:

```json
{
  "sub": "<user uuid>",
  "role": "admin",
  "iss": "duekeep",
  "iat": 0,
  "exp": 0
}
```

## 4) Auth (изменения)

### `POST /api/v1/auth/register`

```json
{ "email": "new@duekeep.local", "password": "secret12" }
```

- `201` пара токенов, `role` = `admin` (в v1 register давал `viewer` — здесь меняем).
- Сразу копия дефолтного дерева категорий (свои UUID), без seed-items.
- Поля `invite_token` нет.
- Занятый email → `409 conflict`.

Тело ответа как login (Sprint 2).

### `GET /api/v1/me`

`200` как Sprint 2: `id`, `email`, `role`. Полей `org_id` / `org_name` нет.

## 5) Org и инвайты

Ручек `/api/v1/org` и `/api/v1/org/invites` нет.

## 6) Предметные ручки

Пути Sprint 3–5 не меняются. Сервер всегда фильтрует по `owner_id` текущего `sub`.

- `GET /items`, `GET /items/{id}`, export/import, renew, bulk — только свои строки.
- `GET/POST/PATCH/DELETE /categories` — только своё дерево (глубина ≤ 3).
- `GET /audit` — только свои события.
- `GET /dashboard`, `GET /calendar`, `GET /notifications` — только свои.
- SSE и Web Push — только клиентам/подпискам с тем же `sub`, что `owner_id` item.
- Чужой UUID предметной сущности → `404 not_found` (не `403`).
- Нельзя передать чужой `owner_id` / `user_id` query/body, чтобы обойти scope.

`GET/POST /kinds` — как Sprint 2 (общий справочник). Кто может писать kinds — [`known-limitations-sprint-7.md`](known-limitations-sprint-7.md).

## 7) Совместимость

- Поля ответов Sprint 2–5 не удалять.
- JWT не расширяем.
- Хендлеры после Sprint 7 не отдают предметные данные без `owner_id` в запросе к БД.
- Клиенты, которые после register ждали `viewer` и общий каталог, больше не верны — это сознательная смена для прода.
