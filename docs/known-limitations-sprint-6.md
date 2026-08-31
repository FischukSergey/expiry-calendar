# Known Limitations — Sprint 6 (v1)

Итог осознанного out of scope. Не баги сдачи.

- нет email / Telegram;
- нет изоляции данных (каталог общий); после v1 свои строки по `owner_id` — [Sprint 7](sprint-7-plan.md), без `org_id`;
- нет вложений, iCal, конвертации валют;
- нет фильтра по JSONB `attrs`;
- нет офлайн-CRUD;
- один инстанс backend (SSE in-memory);
- access после logout жив до `exp` (15 мин);
- Web Push ориентирован на Chromium + localhost;
- после v1: интервал тикера и прочие не-секреты — файл конфига при старте; в `.env` только чувствительные данные (см. [Sprint 4](known-limitations-sprint-4.md)).
