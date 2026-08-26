# Sprint 3 Checklist

Источник: [`sprint-3-plan.md`](sprint-3-plan.md).

## 1) Данные

- [x] Миграция `items` (в т.ч. `attrs`, индексы)
  Примечание: `006_items.sql` — колонки + `attrs` JSONB, `cost_amount INT`, CHECK дат/cost/status, GIN `tags`/`attrs`.
- [x] Миграция `renewals`
  Примечание: `007_renewals.sql` — FK на items (CASCADE) и users.
- [x] Миграция `audit_log`
  Примечание: `008_audit_log.sql` — `actor_id` NULL, индекс `created_at DESC`.
- [x] Часть seed записей можно отложить до Sprint 6; минимум несколько items для тестов
  Примечание: 4 items (rent/subscription/domain/insurance), даты от `Clock.Today()`, статус при записи. Полные 50+ — Sprint 6. `CountItems` больше не заглушка.

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
