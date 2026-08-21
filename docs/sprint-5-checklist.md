# Sprint 5 Checklist

Источник: [`sprint-5-plan.md`](sprint-5-plan.md).

## 1) CSV API

- [ ] `GET /api/v1/items/export` (фильтр как у списка)
- [ ] `POST /api/v1/items/import?dry_run=true`
- [ ] `POST /api/v1/items/import` пишет пачку + audit
- [ ] Маппинг колонок, включая `attrs.*`

## 2) Экраны

- [ ] Вход / выход
- [ ] Дашборд (KPI, bar, pie, топ-10, валюты)
- [ ] Список + фильтры + пагинация
- [ ] Форма создания/редактирования + attrs
- [ ] Карточка + renew + история
- [ ] Календарь месяц
- [ ] Категории
- [ ] Уведомления
- [ ] Аудит (admin)
- [ ] Импорт CSV
- [ ] Viewer без кнопок записи

## 3) Realtime и PWA

- [ ] SSE после логина
- [ ] Разрешение пушей + subscribe
- [ ] Manifest + иконки
- [ ] Service worker, офлайн-заглушка
- [ ] Подсказка установки

## 4) Качество

- [ ] Адаптив desktop/mobile
- [ ] Loading / error / empty
- [ ] [`api-sprint-5.md`](api-sprint-5.md)
- [ ] [`known-limitations-sprint-5.md`](known-limitations-sprint-5.md)
