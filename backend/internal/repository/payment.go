package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"duekeep/internal/db"
	"duekeep/internal/model"
)

// Payments — SQL к item_payments.
type Payments struct {
	pool *pgxpool.Pool
}

// NewPayments создаёт репозиторий оплат вхождения.
func NewPayments(pool *pgxpool.Pool) *Payments {
	return &Payments{pool: pool}
}

func (r *Payments) q(ctx context.Context) db.Querier {
	return db.QuerierFrom(ctx, r.pool)
}

const paymentCols = `id::text, item_id::text, owner_id::text, paid_on, amount, trim(currency), created_at`

// Insert пишет строку. Конфликт (item_id, paid_on) — created=false и существующая строка.
func (r *Payments) Insert(ctx context.Context, p model.ItemPayment) (model.ItemPayment, bool, error) {
	out, err := r.scanOne(ctx, `
INSERT INTO item_payments (item_id, owner_id, paid_on, amount, currency)
VALUES ($1::uuid, $2::uuid, $3, $4, $5)
ON CONFLICT (item_id, paid_on) DO NOTHING
RETURNING `+paymentCols, p.ItemID, p.OwnerID, p.Date, p.Amount, p.Currency)
	if err == nil {
		return out, true, nil
	}
	if !errors.Is(err, model.ErrNotFound) {
		return model.ItemPayment{}, false, err
	}
	existing, gerr := r.GetByItemDate(ctx, p.ItemID, p.Date)
	if gerr != nil {
		return model.ItemPayment{}, false, gerr
	}
	return existing, false, nil
}

// GetByItemDate одна оплата. Нет → ErrNotFound.
func (r *Payments) GetByItemDate(ctx context.Context, itemID, date string) (model.ItemPayment, error) {
	return r.scanOne(ctx, `SELECT `+paymentCols+`
FROM item_payments WHERE item_id = $1::uuid AND paid_on = $2`, itemID, date)
}

// DeleteByItemDate снимает оплату. Нет строки — не ошибка.
func (r *Payments) DeleteByItemDate(ctx context.Context, itemID, date string) error {
	_, err := r.q(ctx).Exec(ctx, `
DELETE FROM item_payments WHERE item_id = $1::uuid AND paid_on = $2`, itemID, date)
	if err != nil {
		return fmt.Errorf("delete payment: %w", err)
	}
	return nil
}

// ListByOwner все оплаты владельца (разреженная таблица).
func (r *Payments) ListByOwner(ctx context.Context, ownerID string) ([]model.ItemPayment, error) {
	if ownerID == "" {
		return []model.ItemPayment{}, nil
	}
	return r.list(ctx, `SELECT `+paymentCols+` FROM item_payments WHERE owner_id = $1::uuid`, ownerID)
}

// ListByItemIDs оплаты выбранных записей (тикер, карточка).
func (r *Payments) ListByItemIDs(ctx context.Context, itemIDs []string) ([]model.ItemPayment, error) {
	if len(itemIDs) == 0 {
		return []model.ItemPayment{}, nil
	}
	return r.list(ctx, `SELECT `+paymentCols+` FROM item_payments WHERE item_id = ANY($1::uuid[])`, itemIDs)
}

func (r *Payments) list(ctx context.Context, q string, args ...any) ([]model.ItemPayment, error) {
	rows, err := r.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()
	out := make([]model.ItemPayment, 0)
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Payments) scanOne(ctx context.Context, q string, args ...any) (model.ItemPayment, error) {
	p, err := scanPayment(r.q(ctx).QueryRow(ctx, q, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ItemPayment{}, model.ErrNotFound
		}
		return model.ItemPayment{}, fmt.Errorf("scan payment: %w", err)
	}
	return p, nil
}

type paymentRow interface {
	Scan(dest ...any) error
}

func scanPayment(row paymentRow) (model.ItemPayment, error) {
	var p model.ItemPayment
	var paidOn time.Time
	if err := row.Scan(&p.ID, &p.ItemID, &p.OwnerID, &paidOn, &p.Amount, &p.Currency, &p.CreatedAt); err != nil {
		return model.ItemPayment{}, err
	}
	p.Date = paidOn.UTC().Format(model.DateLayout)
	return p, nil
}
