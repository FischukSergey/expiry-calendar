# Sprint 3 Checklist

Источник: [`sprint-3-plan.md`](sprint-3-plan.md).

## 1) Данные

- [ ] Миграция `items` (в т.ч. `attrs`, индексы)
- [ ] Миграция `renewals`
- [ ] Миграция `audit_log`
- [ ] Часть seed записей можно отложить до Sprint 6; минимум несколько items для тестов

## 2) API

- [ ] `GET/POST /api/v1/items`
- [ ] `GET/PATCH/DELETE /api/v1/items/{id}`
- [ ] Фильтры и сортировка из контракта
- [ ] Пагинация `page`, `per_page`, `total`
- [ ] `POST /api/v1/items/{id}/renew`
- [ ] `POST /api/v1/items/bulk`
- [ ] `GET /api/v1/audit`
- [ ] Валидация `attrs` по схеме kind
- [ ] Viewer: 403 на мутации

## 3) Аудит

- [ ] create / update / delete / renew / bulk пишут audit
- [ ] before/after без секретов

## 4) Тесты

- [ ] Unit: attrs, статус при записи, глубина уже есть
- [ ] Integration: CRUD, renew, фильтр+page, viewer 403

## 5) DoD

- [ ] [`api-sprint-3.md`](api-sprint-3.md)
- [ ] [`known-limitations-sprint-3.md`](known-limitations-sprint-3.md)
