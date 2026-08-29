-- +goose Up
CREATE TABLE notifications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id    UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    to_status  TEXT NOT NULL CHECK (to_status IN ('expiring', 'expired')),
    title      TEXT NOT NULL,
    read_at    TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- День в UTC, как Clock.Today(). Повтор тикера в тот же день не плодит строки.
CREATE UNIQUE INDEX notifications_item_status_day_uidx
    ON notifications (item_id, to_status, ((created_at AT TIME ZONE 'UTC')::date));

CREATE INDEX notifications_unread_created_idx ON notifications (read_at, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS notifications;
