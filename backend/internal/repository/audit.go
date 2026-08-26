package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"duekeep/internal/db"
	"duekeep/internal/model"
)

// Audit — SQL к audit_log.
type Audit struct {
	pool *pgxpool.Pool
}

// NewAudit создаёт репозиторий журнала.
func NewAudit(pool *pgxpool.Pool) *Audit {
	return &Audit{pool: pool}
}

func (r *Audit) q(ctx context.Context) db.Querier {
	return db.QuerierFrom(ctx, r.pool)
}

// Create пишет событие.
func (r *Audit) Create(ctx context.Context, e model.AuditEntry) error {
	_, err := r.q(ctx).Exec(ctx, `
INSERT INTO audit_log (actor_id, action, entity, entity_id, before_json, after_json)
VALUES (NULLIF($1, '')::uuid, $2, $3, $4::uuid, $5, $6)`,
		actorArg(e.ActorID), e.Action, e.Entity, e.EntityID, nullJSON(e.BeforeJSON), nullJSON(e.AfterJSON))
	if err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}
	return nil
}

// List новые сверху.
func (r *Audit) List(ctx context.Context, page model.Page) ([]model.AuditEntry, int, error) {
	var total int
	if err := r.q(ctx).QueryRow(ctx, `SELECT count(*) FROM audit_log`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit: %w", err)
	}
	rows, err := r.q(ctx).Query(ctx, `
SELECT id::text, actor_id::text, action, entity, entity_id::text, before_json, after_json, created_at
FROM audit_log
ORDER BY created_at DESC, id
LIMIT $1 OFFSET $2`, page.PerPage, page.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("list audit: %w", err)
	}
	defer rows.Close()
	out := make([]model.AuditEntry, 0)
	for rows.Next() {
		var e model.AuditEntry
		if err := rows.Scan(&e.ID, &e.ActorID, &e.Action, &e.Entity, &e.EntityID,
			&e.BeforeJSON, &e.AfterJSON, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}

func actorArg(id *string) string {
	if id == nil {
		return ""
	}
	return *id
}

func nullJSON(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	return raw
}
