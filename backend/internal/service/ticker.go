package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"duekeep/internal/clock"
	"duekeep/internal/model"
)

// TickItemStore — записи, которые тикер может пересчитать.
type TickItemStore interface {
	ListOpen(ctx context.Context) ([]model.Item, error)
	SetStatus(ctx context.Context, id, status string) (model.Item, error)
}

// NotificationStore — лента и идемпотентная вставка за день.
type NotificationStore interface {
	Insert(ctx context.Context, n model.Notification) (model.Notification, bool, error)
	List(ctx context.Context, ownerID string, unread bool, page model.Page) ([]model.Notification, int, error)
	MarkRead(ctx context.Context, id, ownerID string) error
	MarkAllRead(ctx context.Context, ownerID string) error
}

// EventBus — SSE и Web Push (Fanout). nil в тестах, где шина не нужна.
type EventBus interface {
	Notify(n model.Notification)
}

// Ticker пересчитывает статусы по Clock.Today. cancelled/archived/paid не трогает.
type Ticker struct {
	items TickItemStore
	notes NotificationStore
	pays  PaymentStore
	tx    TxFunc
	clk   clock.Clock
	bus   EventBus
}

// NewTicker собирает тикер. Run только в cmd/server; тесты зовут Tick.
func NewTicker(items TickItemStore, notes NotificationStore, tx TxFunc, clk clock.Clock, bus EventBus) *Ticker {
	return &Ticker{items: items, notes: notes, tx: tx, clk: clk, bus: bus}
}

// SetPayments включает overlay дат: порог expiring/expired от ближайшего open.
func (t *Ticker) SetPayments(pays PaymentStore) {
	t.pays = pays
}

// Tick один проход: тот же StatusAtWrite, что при записи. Уведомление — только
// при смене на expiring/expired; повтор в тот же день глотает unique.
func (t *Ticker) Tick(ctx context.Context) error {
	items, err := t.items.ListOpen(ctx)
	if err != nil {
		return err
	}
	today := clock.Today(t.clk)
	now := t.clk.Now().UTC()
	paidIdx, err := t.paymentsByItems(ctx, items)
	if err != nil {
		return err
	}
	for _, it := range items {
		if it.Status == model.StatusPaid {
			continue
		}
		due, ok, err := nextUnpaidOccurrence(it, today, paidDateSet(paidIdx[it.ID]))
		if err != nil {
			return err
		}
		var next string
		if !ok {
			next = model.StatusActive
		} else {
			next = StatusAtWrite(today, due, it.NotifyBeforeDays, "")
		}
		if next == it.Status {
			continue
		}
		item := it
		status := next
		var created model.Notification
		var inserted bool
		err = t.tx(ctx, func(ctx context.Context) error {
			if _, err := t.items.SetStatus(ctx, item.ID, status); err != nil {
				if errors.Is(err, model.ErrNotFound) {
					return nil
				}
				return err
			}
			if item.NotifyBeforeDays == nil {
				return nil
			}
			if status != model.StatusExpiring && status != model.StatusExpired {
				return nil
			}
			var ierr error
			created, inserted, ierr = t.notes.Insert(ctx, model.Notification{
				OwnerID:   item.OwnerID,
				ItemID:    item.ID,
				ToStatus:  status,
				Title:     item.Title,
				CreatedAt: now,
			})
			return ierr
		})
		if err != nil {
			return err
		}
		if inserted && t.bus != nil {
			t.bus.Notify(created)
		}
	}
	return nil
}

func (t *Ticker) paymentsByItems(ctx context.Context, items []model.Item) (map[string]map[string]model.ItemPayment, error) {
	if t.pays == nil || len(items) == 0 {
		return map[string]map[string]model.ItemPayment{}, nil
	}
	ids := make([]string, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.ID)
	}
	rows, err := t.pays.ListByItemIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return paidDatesByItem(rows), nil
}

// Run сразу Tick, затем каждые every. every ≤ 0 — только выход (тесты/выкл).
func (t *Ticker) Run(ctx context.Context, every time.Duration) {
	if every <= 0 {
		return
	}
	if err := t.Tick(ctx); err != nil {
		slog.ErrorContext(ctx, "ticker", "err", err)
	}
	tick := time.NewTicker(every)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if err := t.Tick(ctx); err != nil {
				slog.ErrorContext(ctx, "ticker", "err", err)
			}
		}
	}
}
