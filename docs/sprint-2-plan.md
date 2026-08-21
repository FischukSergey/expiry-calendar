# Sprint 2 — auth и справочники

Источник: [`ARCHITECTURE.md`](../ARCHITECTURE.md) §6–8; [`api-sprint-1.md`](api-sprint-1.md).

## 1) Цель

Вход по JWT + refresh и редактируемые справочники типов и категорий. Без записей истечения.

К концу спринта:

- register / login / refresh / logout / logout-all / me;
- CRUD `item_kinds` (admin) и чтение (viewer);
- CRUD категорий с глубиной ≤ 3 и запретом удалить непустую;
- seed: 2 пользователя, 9 kinds, ≥ 10 категорий.

## 2) Входные условия

Sprint 1 закрыт: compose, `/healthz`, слои, CI.

## 3) Границы

### Входит

- таблицы `users`, `refresh_tokens`, `item_kinds`, `categories`;
- JWT HS256 access 15 мин, opaque refresh 14 дней, ротация, family revoke;
- cookie `duekeep_refresh` + refresh в JSON;
- роли admin / viewer;
- seed пользователей и справочников (идемпотентный).

### Не входит

- `items`, renewals, audit, уведомления;
- UI кроме минимального логина (можно отложить экран в Sprint 5 — тогда в этом спринте только API + curl/Swagger-заглушка не обязательна).
- Решение: **UI логина в Sprint 5**. Здесь проверка через curl / тесты / временный debug не обязателен.

## 4) Backlog

### A. Миграции и repo

- `users`, `refresh_tokens`, `item_kinds`, `categories`.
- repository + service + handler для auth, kinds, categories.

### B. Auth

- bcrypt пароля.
- Claims: `sub`, `role`, `iss=duekeep`, `iat`, `exp`.
- Ротация refresh, reuse → revoke family.
- Middleware Bearer.
- Регистрация всегда `viewer`.

### C. Kinds

- seed 9 slug: domain, subscription, rent, contract, insurance, license, tax, vehicle, other.
- `attr_schema` JSONB.
- DELETE запрещён, если есть items (в Sprint 2 items ещё нет — проверка всё равно пишется).

### D. Categories

- дерево, глубина ≤ 3, цикл запрещён.
- DELETE → 409/202, если дети есть (в Sprint 2 детей-items нет).

### E. Тесты

- login / refresh / logout / reuse family;
- viewer не может POST kind;
- глубина категории.

## 5) Техрешения

- Интерфейсы repo объявляет **service**.
- Access не в БД. Refresh — только `token_hash`.
- Cookie Path=`/api/v1/auth`.
- Seed по email и slug, повторный up не плодит строки.

## 6) DoD

- admin и viewer логинятся, получают пару токенов;
- refresh крутит пару, старый refresh 401;
- reuse отозванного/старого в family даёт 401 на всю family;
- viewer читает kinds/categories, не пишет;
- контракт [`api-sprint-2.md`](api-sprint-2.md).

## 7) Демо

1. `POST /auth/login` admin@duekeep.local.
2. `GET /me`, `GET /kinds`, `GET /categories`.
3. `POST /auth/refresh`.
4. Логин viewer, `POST /kinds` → 403.

## 8) Риски

- расхождение cookie vs body refresh → принимать оба, один источник истины в service.
- seed паролей в README, не в git как прод-секрет.

## 9) Артефакты

- миграции и seed справочников;
- auth middleware;
- тесты токенов;
- docs спринта.
