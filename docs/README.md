# Документация спринтов Duekeep

Источник истины по системе: [`../ARCHITECTURE.md`](../ARCHITECTURE.md), [`../FUNCTIONAL.md`](../FUNCTIONAL.md).

Каждый спринт — отдельный коммит (или пачка связанных), не весь проект разом.

| Спринт | Тема | План | API | Чеклист | Ограничения |
|---|---|---|---|---|---|
| 1 | Каркас | [plan](sprint-1-plan.md) | [api](api-sprint-1.md) | [checklist](sprint-1-checklist.md) | [limitations](known-limitations-sprint-1.md) |
| 2 | Auth и справочники | [plan](sprint-2-plan.md) | [api](api-sprint-2.md) | [checklist](sprint-2-checklist.md) | [limitations](known-limitations-sprint-2.md) |
| 3 | Записи | [plan](sprint-3-plan.md) | [api](api-sprint-3.md) | [checklist](sprint-3-checklist.md) | [limitations](known-limitations-sprint-3.md) |
| 4 | Статусы, realtime, обзор | [plan](sprint-4-plan.md) | [api](api-sprint-4.md) | [checklist](sprint-4-checklist.md) | [limitations](known-limitations-sprint-4.md) |
| 5 | UI, CSV, PWA | [plan](sprint-5-plan.md) | [api](api-sprint-5.md) | [checklist](sprint-5-checklist.md) | [limitations](known-limitations-sprint-5.md) |
| 6 | Сдача | [plan](sprint-6-plan.md) | [api](api-sprint-6.md) | [checklist](sprint-6-checklist.md) | [limitations](known-limitations-sprint-6.md) |
| 7 | Изоляция данных (org) | [plan](sprint-7-plan.md) | [api](api-sprint-7.md) | [checklist](sprint-7-checklist.md) | [limitations](known-limitations-sprint-7.md) |

Правило: новый handler не меняет контракт спринта без правки `api-sprint-N.md`.

Код не пишем, пока не начат Sprint 1 по чеклисту. Sprint 7 — после закрытия v1 (спринты 1–6).
