# Sprint 5 — UI, CSV, PWA

Источник: [`FUNCTIONAL.md`](../FUNCTIONAL.md) экраны; [`ARCHITECTURE.md`](../ARCHITECTURE.md) §10–11.

## 1) Цель

Пользователь работает в браузере и установленном PWA: все экраны, CSV, пуши, адаптив.

## 2) Входные условия

Sprint 4: все нужные API уже есть, кроме CSV.

## 3) Границы

### Входит

- CSV import (маппинг + dry_run) и export текущего фильтра;
- экраны: логин, дашборд, список, форма, карточка, календарь, категории, уведомления, аудит, импорт, профиль;
- access в памяти, refresh по cookie / 401 interceptor;
- EventSource после логина;
- разрешение Notification + subscribe;
- PWA: manifest, иконки 192/512, Workbox, офлайн-заглушка, beforeinstallprompt;
- адаптив (сайдбар / нижние табы).

### Не входит

- офлайн-CRUD;
- живой Swagger (Sprint 6, но UI уже ходит в те же контракты);
- полный seed 50+ (можно черновой, добить в 6).

## 4) Backlog

### A. CSV backend + тонкий UI маппинга колонок.

### B. React Router, TanStack Query, zod-формы, Recharts.

### C. Auth interceptor + SSE hook + push hook.

### D. PWA vite-plugin-pwa.

### E. Состояния loading / empty / error, мобильная вёрстка.

## 5) Техрешения

- `credentials: 'include'` только auth refresh/logout.
- Форма item: общие поля + блок из `attr_schema`.
- Export бьёт тот же query, что список.

## 6) DoD

- пройти демо-сценарий из FUNCTIONAL без Swagger;
- PWA ставится в Chrome на localhost;
- viewer не видит кнопок мутаций (только 403 недостаточно);
- [`api-sprint-5.md`](api-sprint-5.md) (CSV).

## 7) Демо

Логин admin, дашборд, фильтр, карточка, renew, календарь, CSV туда-обратно, колокольчик, «Установить».

## 8) Риски

- SW кэширует старый index → network-first для HTML.
- EventSource и истекший access → пересоздать после refresh.

## 9) Артефакты

- SPA/PWA, CSV API, docs.
