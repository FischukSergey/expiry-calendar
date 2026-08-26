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

- [x] `GET/POST /api/v1/items`
  Примечание: GET под Bearer; POST admin, 201, status уже посчитан. Дефолты RUB / one_time / 30 дней.
- [x] `GET/PATCH/DELETE /api/v1/items/{id}`
  Примечание: GET → `{item, renewals}`; PATCH/DELETE — admin, 200/204.
- [x] Фильтры и сортировка из контракта
  Примечание: q (title/vendor/tags), kind/status/category+потомки (CTE), vendor, даты, цена, billing, tag; sort whitelist.
- [x] Пагинация `page`, `per_page`, `total`
  Примечание: дефолт 1/20, per_page max 100.
- [x] `POST /api/v1/items/{id}/renew`
  Примечание: пишет `renewals` + audit; `bulk` зарегистрирован до `/{id}`.
- [x] `POST /api/v1/items/bulk`
  Примечание: ids + хотя бы одно из category_id/status (cancelled|archived).
- [x] `GET /api/v1/audit`
  Примечание: admin, пагинация как у items.
- [x] Валидация `attrs` по схеме kind
  Примечание: лишние ключи и тип → 422; пустая схема → `{}`.
- [x] Viewer: 403 на мутации
  Примечание: `RequireAdmin` на POST/PATCH/DELETE/renew/bulk и GET /audit. GET items — viewer 200.

## 3) Аудит

- [x] create / update / delete / renew / bulk пишут audit
  Примечание: в той же tx, что мутация. `TestMutationsWriteAudit` — все пять `action`.
- [x] before/after без секретов
  Примечание: `itemAuditSnap` — id/title/kind/category/status/expires/cost/attrs. Нет url, account_hint, паролей.

## 4) Тесты

- [x] Unit: attrs, статус при записи, глубина уже есть
  Примечание: `TestValidateAttrs`, `TestStatusAtWrite`, `TestCategoryDepthAndCreateLimit`.
- [x] Integration: CRUD, renew, фильтр+page, viewer 403
  Примечание: `TestItemsCRUDRenewFilterPage` — patch/renew+история, q+page, tag. Viewer 403 на POST/PATCH/DELETE/renew/bulk и GET /audit.

## 5) DoD

- [x] [`api-sprint-3.md`](api-sprint-3.md)
  Примечание: ручки items/audit совпадают с хендлерами и `openapi.yaml`. Дефолты и снимок audit зафиксированы.
- [x] [`known-limitations-sprint-3.md`](known-limitations-sprint-3.md)
  Примечание: нет тикера, нет UI/SSE/dashboard, seed 50+ в Sprint 6, audit без фильтров, integration без Postgres.
