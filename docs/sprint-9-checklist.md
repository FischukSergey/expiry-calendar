# Sprint 9 Checklist

Источник: [`sprint-9-plan.md`](sprint-9-plan.md). Код — только когда пользователь попросил реализовать спринт. Документы можно завести заранее.

## 1) Схема

- [x] Миграция: `status` CHECK включает `paid`. (`012_paid_notify.sql`)
- [x] `notify_before_days` допускает `NULL` («не уведомлять»). `0` = в день срока.
- [x] Старые строки без `NULL` и без авто-`paid`. (только DROP NOT NULL, данные не трогаем)

## 2) Статус «Оплачено»

- [x] Константа `paid`, бейдж «Оплачено», фильтр списка.
- [x] Форма / PATCH / bulk: можно выставить `paid`; тикер `paid` не пересчитывает и не создаёт notification (предоплата).
- [x] Текущее вхождение `paid` не в «сгорит» / графике / календаре; следующие циклы monthly/yearly — да (не как полное исключение `cancelled`).
- [x] `renew` с `paid` разрешён; после смены даты статус снова считает тикер.

## 3) Тип «Мобильная связь»

- [x] Seed / `EnsureKinds`: slug `mobile`, name «Мобильная связь», стабильный UUID, ON CONFLICT по slug.
- [x] `CheckCatalog` / `requiredKindSlugs` — 10 типов, включая `mobile`.
- [x] В селекте формы записи тип есть. Экрана и кнопки «добавить тип» нет.

## 4) Не уведомлять

- [x] JSON `notify_before_days: null`; форма — чекбокс, поле дней выключается.
- [x] Тикер: нет notification / SSE / push; нет перехода в `expiring`.
- [x] CSV: пусто или `off` → `null`.

## 5) Обзор и календарь: периоды

- [x] `expirations_by_month` и `GET /calendar` разворачивают `monthly` / `yearly` от якоря `expires_at` (clamp 29–31).
- [x] `one_time` без изменений. `upcoming_cost` как Sprint 4 (run-rate).
- [x] `paid`: текущее вхождение (`expires_at`) не в графике/календаре/«сгорит»; следующие циклы — да.
- [x] Тест: monthly виден не только в месяце `expires_at`; `paid` прячет текущий день.

## 6) Лента PWA

- [x] Нижний `nav` (`Layout`, `lg:hidden`): шрифт подписей крупнее `11px` (ориентир `text-sm`).
- [x] Пять вкладок и safe-area: текст не обрезан. Десктопное меню без обязательных правок.

## 7) Спека и тесты

- [x] [`api-sprint-9.md`](api-sprint-9.md) и `backend/openapi.yaml`.
- [x] Тесты: `paid` + тикер (статус и нет notification); `null` notify + тикер; create/PATCH; seed содержит `mobile`; развёртка monthly.

## 8) DoD

- [x] [`known-limitations-sprint-9.md`](known-limitations-sprint-9.md) заполнен.
- [x] Демо плана §7: API на локальном compose (kinds=10/`mobile`, create `paid` и `notify=null`, monthly на календаре соседнего месяца, колокольчик без Sprint9). Ленту PWA в браузере не кликали — в бандле есть `text-sm leading-tight`, «Оплачено», «Не уведомлять».
- [x] `task lint` и `task test` зелёные.
