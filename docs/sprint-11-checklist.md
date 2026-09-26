# Sprint 11 Checklist

Источник: [`sprint-11-plan.md`](sprint-11-plan.md).

## 1) Просрочка и календарный шаг

- [x] Статус, `next_open_at`, soonest и «оплатить ближайшее» смотрят на самое раннее неоплаченное вхождение, в том числе раньше today.
- [x] Неоплаченный день `< today` → `expired`. Платёж на этот день переносит якорь. Заморозка `paid` и `cancelled` / `archived` не пересчитываются.
- [x] Обход месяцев от 1-го числа + `clampDay`. Тест: 31 января не пропускает 28 февраля.
- [x] `expiring_7` / `30` не считают прошлые даты. Soonest сортирует просроченные первыми.
- [x] Тест: monthly с вчерашним open и будущим периодом остаётся `expired`, пока вчера не оплачено.
  Примечание: `TestMonthlyPastUnpaidStaysExpiredUntilPaid`, `TestNextUnpaidKeepsFebruaryWhenTodayIs31`. Карточка на стенде: якорь 01.07.2026, бейдж «Просрочено», «просрочено на 87 дн.».

## 2) Refresh reuse

- [x] `RevokeFamily` при reuse коммитится, клиент получает 401, новая пара не выдаётся.
- [x] Fake-транзакция откатывает снимок при ошибке колбэка. Тест семьи на этом fake красный, если revoke внутри отката.
- [x] `task test:integration`: compose `deploy/test`, тег `integration`, сценарий login → refresh → replay → новый токен тоже 401.
- [x] CI job гоняет тот же тег на service Postgres. `task test` тег не включает.

## 3) Роль администратора справочника

- [x] Миграция: `users.role` допускает `administrator`. Существующие пользователи остаются `admin`.
- [x] `POST/PATCH/DELETE /kinds` только у `administrator`. `admin` и `viewer` → 403. `GET /kinds` любому auth.
- [x] `Register` создаёт `admin`. Тест регистрации: свой item можно, kind нельзя. Отдельный тест на роль `administrator`.
- [x] `ADMINISTRATOR_EMAIL`: старт и register/login повышают этот email. Пустой env — справочник только для чтения. Публичной смены роли нет.
- [x] `ARCHITECTURE.md` §7 и `FUNCTIONAL.md` описывают три роли и отсутствие демо-каталога.

## 4) Удаление seed

- [x] Нет `seed.Run`, пакета демо-фикстур и env `SEED` в процессе и compose.
- [x] Десять типов перенесены в goose (`ON CONFLICT DO NOTHING`). `EnsureKinds` не вызывается.
  Примечание: локальный стенд накатил `014_administrator.sql`.
- [x] Шаблон категорий при регистрации работает из нового места. Пакет `internal/seed` удалён.
- [x] README без паролей `admin@` / `viewer@` и без описания демо-каталога. Миграция не удаляет уже лежащие в volume строки.

## 5) UI

- [x] Карточка: отсчёт до `next_open_at` (просрочено / сегодня / через N дней). Нет open — блока нет.
- [x] Форма не отправляет `started_at` позже `expires_at`.
  Примечание: форма показала «не позже даты вхождения» и осталась на `/items/new`.
- [x] Список: admin выбирает строки и меняет категорию и/или статус через `POST /items/bulk`. Viewer без чекбоксов.
  Примечание: чекбокс и панель «Выбрано: 1» на списке. Viewer по-прежнему без `isAdmin`.

## 6) Сеть, лог, экспорт

- [x] Push: только публичный `https`. Частный и loopback → 422. Перед отправкой проверка повторяется, внутренний адрес не запрашивается.
- [x] 500 логируются через один helper, клиенту текст ошибки не отдаётся.
- [x] Login: 8 неудач / 15 минут на IP+email → 429. Успех сбрасывает ключ.
  Примечание: `TestLoginRateLimit`, девятая попытка `429` `rate_limited`.
- [x] CSV: ячейка с `=`, `+`, `-`, `@`, табом или CR получает префикс `'`.

## 7) Спека и тесты

- [x] [`api-sprint-11.md`](api-sprint-11.md) и `backend/openapi.yaml`.
- [x] Vitest: текст отсчёта и порядок дат. `npm test` подключён к `task test:frontend` и к CI frontend.
- [x] [`known-limitations-sprint-11.md`](known-limitations-sprint-11.md) заполнен по факту сдачи.
- [x] `REPORT.md` дополнен по ходу.

## 8) DoD

- [x] Демо плана §7.
  Примечание: регистрация без демо-учётки, kinds на месте, карточка просрочена до оплаты июля, форма дат, bulk-панель. `POST /kinds` для admin — тесты 403.
- [x] `task lint`, `task test`, `task test:integration` зелёные.
