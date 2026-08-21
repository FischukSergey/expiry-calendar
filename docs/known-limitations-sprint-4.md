# Known Limitations — Sprint 4

## Realtime

**Один процесс backend.** Hub в памяти. Второй инстанс не получит SSE.

**Пропущенные SSE при офлайне не буфер.** После reconnect — `GET /notifications`.

## Push

**Демо: Chrome + localhost.** iOS только у установленного PWA, не цель защиты.

**Нет UI подписки.** API есть, кнопка разрешения — Sprint 5.

## UI

Дашборд и календарь есть только как JSON.

## Тикер

Интервал 60 с; в тестах — явный tick, не ждать минуту.
