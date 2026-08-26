-- +goose Up
CREATE TABLE renewals (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id        UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    actor_id       UUID NOT NULL REFERENCES users(id),
    old_expires_at DATE NOT NULL,
    new_expires_at DATE NOT NULL,
    old_cost       INT NOT NULL,
    new_cost       INT NOT NULL,
    comment        TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS renewals;
