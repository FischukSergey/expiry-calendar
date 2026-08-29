package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"duekeep/internal/db"
	"duekeep/internal/model"
)

const pushCols = `id::text, user_id::text, endpoint, p256dh, auth, user_agent, created_at`

// PushSubscriptions — SQL к push_subscriptions.
type PushSubscriptions struct {
	pool *pgxpool.Pool
}

// NewPushSubscriptions создаёт репозиторий Web Push подписок.
func NewPushSubscriptions(pool *pgxpool.Pool) *PushSubscriptions {
	return &PushSubscriptions{pool: pool}
}

func (r *PushSubscriptions) q(ctx context.Context) db.Querier {
	return db.QuerierFrom(ctx, r.pool)
}

// Upsert пишет или обновляет строку по endpoint (устройство могло сменить ключи).
func (r *PushSubscriptions) Upsert(ctx context.Context, s model.PushSubscription) error {
	_, err := r.q(ctx).Exec(ctx, `
INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth, user_agent)
VALUES ($1::uuid, $2, $3, $4, $5)
ON CONFLICT (endpoint) DO UPDATE SET
    user_id = EXCLUDED.user_id,
    p256dh = EXCLUDED.p256dh,
    auth = EXCLUDED.auth,
    user_agent = EXCLUDED.user_agent`,
		s.UserID, s.Endpoint, s.P256dh, s.Auth, s.UserAgent)
	if err != nil {
		return fmt.Errorf("upsert push subscription: %w", err)
	}
	return nil
}

// DeleteByEndpoint снимает подписку. Нет строки — не ошибка.
func (r *PushSubscriptions) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	_, err := r.q(ctx).Exec(ctx, `DELETE FROM push_subscriptions WHERE endpoint = $1`, endpoint)
	if err != nil {
		return fmt.Errorf("delete push subscription: %w", err)
	}
	return nil
}

// List все подписки: данные общие, пуш уходит каждому подписанному устройству.
func (r *PushSubscriptions) List(ctx context.Context) ([]model.PushSubscription, error) {
	rows, err := r.q(ctx).Query(ctx, `SELECT `+pushCols+` FROM push_subscriptions`)
	if err != nil {
		return nil, fmt.Errorf("list push subscriptions: %w", err)
	}
	defer rows.Close()
	out := make([]model.PushSubscription, 0)
	for rows.Next() {
		var s model.PushSubscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.Endpoint, &s.P256dh, &s.Auth, &s.UserAgent, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
