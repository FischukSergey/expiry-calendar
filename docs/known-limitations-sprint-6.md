# Known Limitations — Sprint 6 (v1)

Итог осознанного out of scope. Не баги сдачи.

- нет email / Telegram;
- нет `org_id` и изоляции данных (токены уже многопользовательские);
- нет вложений, iCal, конвертации валют;
- нет фильтра по JSONB `attrs`;
- нет офлайн-CRUD;
- один инстанс backend (SSE in-memory);
- access после logout жив до `exp` (15 мин);
- Web Push ориентирован на Chromium + localhost.
