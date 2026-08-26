package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"duekeep/internal/db"
	"duekeep/internal/model"
)

// Renewals — SQL к renewals.
type Renewals struct {
	pool *pgxpool.Pool
}

// NewRenewals создаёт репозиторий продлений.
func NewRenewals(pool *pgxpool.Pool) *Renewals {
	return &Renewals{pool: pool}
}

func (r *Renewals) q(ctx context.Context) db.Querier {
	return db.QuerierFrom(ctx, r.pool)
}

// Create пишет строку истории.
func (r *Renewals) Create(ctx context.Context, rec model.Renewal) (model.Renewal, error) {
	var out model.Renewal
	var oldExp, newExp time.Time
	err := r.q(ctx).QueryRow(ctx, `
INSERT INTO renewals (
    item_id, actor_id, old_expires_at, new_expires_at, old_cost, new_cost, comment
) VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7)
RETURNING id::text, item_id::text, actor_id::text, old_expires_at, new_expires_at,
          old_cost, new_cost, comment, created_at`,
		rec.ItemID, rec.ActorID, rec.OldExpiresAt, rec.NewExpiresAt, rec.OldCost, rec.NewCost, rec.Comment,
	).Scan(&out.ID, &out.ItemID, &out.ActorID, &oldExp, &newExp,
		&out.OldCost, &out.NewCost, &out.Comment, &out.CreatedAt)
	if err != nil {
		return model.Renewal{}, fmt.Errorf("insert renewal: %w", err)
	}
	out.OldExpiresAt = oldExp.UTC().Format(model.DateLayout)
	out.NewExpiresAt = newExp.UTC().Format(model.DateLayout)
	return out, nil
}

// ListByItem история по записи, новые сверху.
func (r *Renewals) ListByItem(ctx context.Context, itemID string) ([]model.Renewal, error) {
	rows, err := r.q(ctx).Query(ctx, `
SELECT id::text, item_id::text, actor_id::text, old_expires_at, new_expires_at,
       old_cost, new_cost, comment, created_at
FROM renewals WHERE item_id = $1::uuid
ORDER BY created_at DESC`, itemID)
	if err != nil {
		return nil, fmt.Errorf("list renewals: %w", err)
	}
	defer rows.Close()
	out := make([]model.Renewal, 0)
	for rows.Next() {
		var rec model.Renewal
		var oldExp, newExp time.Time
		if err := rows.Scan(&rec.ID, &rec.ItemID, &rec.ActorID, &oldExp, &newExp,
			&rec.OldCost, &rec.NewCost, &rec.Comment, &rec.CreatedAt); err != nil {
			return nil, err
		}
		rec.OldExpiresAt = oldExp.UTC().Format(model.DateLayout)
		rec.NewExpiresAt = newExp.UTC().Format(model.DateLayout)
		out = append(out, rec)
	}
	return out, rows.Err()
}
