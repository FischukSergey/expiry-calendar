# Sprint 5 Checklist

Источник: [`sprint-5-plan.md`](sprint-5-plan.md).

## 1) CSV API

- [x] `GET /api/v1/items/export` (фильтр как у списка)
  Примечание: те же query, без page/per_page; потолок 10_000. Viewer 200, `text/csv`. `TestExportFilterAndAttrs`.
- [x] `POST /api/v1/items/import?dry_run=true`
  Примечание: multipart `file`+`mapping`; не пишет БД и audit. `TestImportDryRunDoesNotWrite`.
- [x] `POST /api/v1/items/import` пишет пачку + audit
  Примечание: одна транзакция; любая ошибка строки → 422, ничего не пишем. Audit `import`. `TestImportWritesBatchAndAudit`.
- [x] Маппинг колонок, включая `attrs.*`
  Примечание: ключ поля → имя колонки CSV. `NormalizeCSVMapping`, импорт `attrs.registrar`.

## 2) Экраны

- [x] Вход / выход
  Примечание: login/register, access в памяти, refresh по cookie, 401 interceptor. Профиль — выход и logout-all.
- [x] Дашборд (KPI, bar, pie, топ-10, валюты)
  Примечание: Recharts; pie с переключателем валюты; upcoming_cost раздельно.
- [x] Список + фильтры + пагинация
  Примечание: query в URL; экспорт текущего фильтра.
- [x] Форма создания/редактирования + attrs
  Примечание: блок из attr_schema; пресеты 7/14/30.
- [x] Карточка + renew + история
- [x] Календарь месяц
  Примечание: сетка, точки по статусу, клик — список дня.
- [x] Категории
  Примечание: дерево; мутации только admin.
- [x] Уведомления
- [x] Аудит (admin)
- [x] Импорт CSV
  Примечание: маппинг, dry_run, запись.
- [x] Viewer без кнопок записи
  Примечание: кнопки скрыты по роли, не только 403. Аудит/импорт/форма — AdminRoute.

## 3) Realtime и PWA

- [x] SSE после логина
  Примечание: EventSource `?access_token=`; на notification инвалидация ленты/списка/дашборда/календаря; после refresh — новое соединение.
- [x] Разрешение пушей + subscribe
  Примечание: после входа `Notification` + `pushManager.subscribe` + `POST /push/subscribe`. Профиль: разрешить / отписаться.
- [x] Manifest + иконки
  Примечание: Duekeep, standalone, 192/512 PNG.
- [x] Service worker, офлайн-заглушка
  Примечание: injectManifest + Workbox; HTML/API network-first; SSE не кэшируем; `offline.html`. Клик пуша → `/items/:id`.
- [x] Подсказка установки
  Примечание: `beforeinstallprompt` — баннер в layout и кнопка в профиле.

## 4) Качество

- [x] Адаптив desktop/mobile
  Примечание: сайдбар / нижние табы + safe-area; список карточками на узком экране; календарь компактный; таблицы с горизонтальным скроллом.
- [x] Loading / error / empty
  Примечание: `PageState` на экранах; ошибка — «Повторить»; импорт без файла и пустые списки.
- [x] [`api-sprint-5.md`](api-sprint-5.md)
  Примечание: export/import сверены с хендлерами и OpenAPI (`exportItems`, `importItems`).
- [x] [`known-limitations-sprint-5.md`](known-limitations-sprint-5.md)
  Примечание: нет офлайн-CRUD, seed 50+ в Sprint 6, нет UI kinds/bulk.

## 5) DoD

- [x] Демо-сценарий без Swagger
  Примечание: login → dashboard (RUB+USD) → список/карточка → calendar → export CSV → import dry_run. Viewer: чтение 200, мутации 403, кнопок записи в UI нет.
- [x] PWA на localhost
  Примечание: manifest, sw.js no-cache, иконки 192/512, offline.html. Установка — Chrome, часто со второго захода.
- [x] `task lint` / `task test` зелёные
