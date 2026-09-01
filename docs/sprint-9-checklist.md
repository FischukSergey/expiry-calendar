# Sprint 9 Checklist

Источник: [`sprint-9-plan.md`](sprint-9-plan.md). Код — только когда пользователь попросил реализовать спринт. Документы можно завести заранее.

## 1) Схема

- [ ] Миграция: `status` CHECK включает `paid`.
- [ ] `notify_before_days` допускает `NULL` («не уведомлять»). `0` = в день срока.
- [ ] Старые строки без `NULL` и без авто-`paid`.

## 2) Статус «Оплачено»

- [ ] Константа `paid`, бейдж «Оплачено», фильтр списка.
- [ ] Форма / PATCH / bulk: можно выставить `paid`; тикер `paid` не пересчитывает и не создаёт notification (предоплата).
- [ ] Дашборд и календарь открытых без `paid` (как без `cancelled`).
- [ ] `renew` с `paid` разрешён; после смены даты статус снова считает тикер.

## 3) Тип «Мобильная связь»

- [ ] Seed / `EnsureKinds`: slug `mobile`, name «Мобильная связь», стабильный UUID, ON CONFLICT по slug.
- [ ] `CheckCatalog` / `requiredKindSlugs` — 10 типов, включая `mobile`.
- [ ] В селекте формы записи тип есть. Экрана и кнопки «добавить тип» нет.

## 4) Не уведомлять

- [ ] JSON `notify_before_days: null`; форма — чекбокс, поле дней выключается.
- [ ] Тикер: нет notification / SSE / push; нет перехода в `expiring`.
- [ ] CSV: пусто или `off` → `null`.

## 5) Лента PWA

- [ ] Нижний `nav` (`Layout`, `lg:hidden`): шрифт подписей крупнее `11px` (ориентир `text-sm`).
- [ ] Пять вкладок и safe-area: текст не обрезан. Десктопное меню без обязательных правок.

## 6) Спека и тесты

- [ ] [`api-sprint-9.md`](api-sprint-9.md) и `backend/openapi.yaml`.
- [ ] Тесты: `paid` + тикер (статус и нет notification); `null` notify + тикер; create/PATCH; seed содержит `mobile`.

## 7) DoD

- [ ] [`known-limitations-sprint-9.md`](known-limitations-sprint-9.md) заполнен.
- [ ] Демо плана §7 пройдено.
- [ ] `task lint` и `task test` зелёные.
