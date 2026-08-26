# Known Limitations — Sprint 3

## Статусы

**Нет фонового тикера.** Запись, созданная вчера как `active`, не станет `expiring` overnight без PATCH или до Sprint 4.

## Уведомления

Нет in-app, SSE, Web Push.

## Обзор

Нет `/dashboard`, `/calendar`, CSV, UI.

## Seed

Полные 50+ записей — Sprint 6. В Sprint 3 достаточно фикстур для тестов.

## JSONB

Фильтра по ключам `attrs` нет и не планируется в v1.

## Цены

`cost_amount` и `new_cost` — целые (валюта без копеек/центов). Дробную часть не принимаем.

## Audit

**GET `/audit` без фильтров** по actor/entity/action. Только `page` / `per_page`.

## Тесты

**Нет integration с Postgres.** CRUD, renew, фильтр+page, viewer 403 — httptest + память. `task test:integration` по-прежнему заглушка.
