package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"duekeep/internal/db"
	"duekeep/internal/model"
)

const notificationCols = `id::text, owner_id::text, item_id::text, to_status, title, read_at, created_at`

// Notifications — SQL к notifications.
type Notifications struct {
	pool *pgxpool.Pool
}

// NewNotifications создаёт репозиторий уведомлений.
func NewNotifications(pool *pgxpool.Pool) *Notifications {
	return &Notifications{pool: pool}
}

func (r *Notifications) q(ctx context.Context) db.Querier {
	return db.QuerierFrom(ctx, r.pool)
}

// Insert пишет строку. Конфликт (item, to_status, день UTC) — false, не ошибка.
func (r *Notifications) Insert(ctx context.Context, n model.Notification) (model.Notification, bool, error) {
	created, err := scanNotification(r.q(ctx).QueryRow(ctx, `
INSERT INTO notifications (owner_id, item_id, to_status, title, created_at)
VALUES ($1::uuid, $2::uuid, $3, $4, $5)
ON CONFLICT (item_id, to_status, ((created_at AT TIME ZONE 'UTC')::date)) DO NOTHING
RETURNING `+notificationCols, n.OwnerID, n.ItemID, n.ToStatus, n.Title, n.CreatedAt))
	if errors.Is(err, model.ErrNotFound) {
		return model.Notification{}, false, nil
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Notification{}, false, nil
		}
		return model.Notification{}, false, fmt.Errorf("insert notification: %w", err)
	}
	return created, true, nil
}

// List новые сверху, только своего владельца. unread — только без read_at.
func (r *Notifications) List(
	ctx context.Context, ownerID string, unread bool, page model.Page,
) ([]model.Notification, int, error) {
	if ownerID == "" {
		return []model.Notification{}, 0, nil
	}
	where := " WHERE owner_id = $1::uuid"
	args := make([]any, 0, 3)
	args = append(args, ownerID)
	if unread {
		where += " AND read_at IS NULL"
	}
	var total int
	if err := r.q(ctx).QueryRow(ctx, `SELECT count(*) FROM notifications`+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}
	lim := len(args) + 1
	off := len(args) + 2
	args = append(args, page.PerPage, page.Offset())
	rows, err := r.q(ctx).Query(ctx, fmt.Sprintf(`SELECT %s
FROM notifications%s
ORDER BY created_at DESC, id
LIMIT $%d OFFSET $%d`, notificationCols, where, lim, off), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()
	out := make([]model.Notification, 0)
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, n)
	}
	return out, total, rows.Err()
}

// MarkRead ставит read_at. Повтор — 204. Чужой или нет строки → ErrNotFound.
func (r *Notifications) MarkRead(ctx context.Context, id, ownerID string) error {
	if ownerID == "" {
		return model.ErrNotFound
	}
	tag, err := r.q(ctx).Exec(ctx, `
UPDATE notifications SET read_at = COALESCE(read_at, now())
WHERE id = $1::uuid AND owner_id = $2::uuid`, id, ownerID)
	if err != nil {
		return fmt.Errorf("read notification: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

// MarkAllRead помечает непрочитанные владельца. Пустой набор — не ошибка.
func (r *Notifications) MarkAllRead(ctx context.Context, ownerID string) error {
	if ownerID == "" {
		return nil
	}
	_, err := r.q(ctx).Exec(ctx, `
UPDATE notifications SET read_at = now() WHERE read_at IS NULL AND owner_id = $1::uuid`, ownerID)
	if err != nil {
		return fmt.Errorf("read all notifications: %w", err)
	}
	return nil
}

type notificationRow interface {
	Scan(dest ...any) error
}

func scanNotification(row notificationRow) (model.Notification, error) {
	var n model.Notification
	var readAt *time.Time
	if err := row.Scan(&n.ID, &n.OwnerID, &n.ItemID, &n.ToStatus, &n.Title, &readAt, &n.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Notification{}, model.ErrNotFound
		}
		return model.Notification{}, err
	}
	n.ReadAt = readAt
	return n, nil
}
