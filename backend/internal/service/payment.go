package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"duekeep/internal/model"
)

const fieldDate = "date"

// PaymentStore — разреженный журнал оплат вхождения.
type PaymentStore interface {
	Insert(ctx context.Context, p model.ItemPayment) (model.ItemPayment, bool, error)
	GetByItemDate(ctx context.Context, itemID, date string) (model.ItemPayment, error)
	DeleteByItemDate(ctx context.Context, itemID, date string) error
	ListByOwner(ctx context.Context, ownerID string) ([]model.ItemPayment, error)
	ListByItemIDs(ctx context.Context, itemIDs []string) ([]model.ItemPayment, error)
}

// SetPayments включает журнал вхождений. Без вызова Pay/Unpay и next_open_at не работают.
func (s *Item) SetPayments(pays PaymentStore) {
	s.pays = pays
}

// Pay отмечает оплату даты. Повтор — та же строка, created=false (HTTP 200).
func (s *Item) Pay(ctx context.Context, id, date, actorID string) (model.ItemPayment, bool, error) {
	if s.pays == nil {
		return model.ItemPayment{}, false, model.ErrInternal
	}
	it, err := s.items.ByID(ctx, id)
	if err != nil {
		return model.ItemPayment{}, false, err
	}
	if err := requireOwner(it.OwnerID, actorID); err != nil {
		return model.ItemPayment{}, false, err
	}
	day, err := parseDate(fieldDate, date)
	if err != nil {
		return model.ItemPayment{}, false, err
	}
	ok, err := isOccurrenceDate(it, day)
	if err != nil {
		return model.ItemPayment{}, false, err
	}
	if !ok {
		return model.ItemPayment{}, false, model.Validation("date is not an occurrence", map[string]any{fieldDate: "not in series"})
	}
	in := model.ItemPayment{
		ItemID: it.ID, OwnerID: it.OwnerID, Date: day.Format(model.DateLayout),
		Amount: it.CostAmount, Currency: it.Currency,
	}
	var out model.ItemPayment
	var created bool
	err = s.tx(ctx, func(ctx context.Context) error {
		var ierr error
		out, created, ierr = s.pays.Insert(ctx, in)
		if ierr != nil {
			return ierr
		}
		if !created {
			return nil
		}
		return s.audit.Create(ctx, auditEntry(actorID, model.AuditPay, it.ID, nil, paymentSnap(out)))
	})
	return out, created, err
}

// Unpay снимает оплату даты. Нет строки — успех без audit.
func (s *Item) Unpay(ctx context.Context, id, date, actorID string) error {
	if s.pays == nil {
		return model.ErrInternal
	}
	it, err := s.items.ByID(ctx, id)
	if err != nil {
		return err
	}
	if err := requireOwner(it.OwnerID, actorID); err != nil {
		return err
	}
	day, err := parseDate(fieldDate, date)
	if err != nil {
		return err
	}
	key := day.Format(model.DateLayout)
	return s.tx(ctx, func(ctx context.Context) error {
		cur, err := s.pays.GetByItemDate(ctx, it.ID, key)
		if err != nil {
			if errors.Is(err, model.ErrNotFound) {
				return nil
			}
			return err
		}
		if err := s.pays.DeleteByItemDate(ctx, it.ID, key); err != nil {
			return err
		}
		return s.audit.Create(ctx, auditEntry(actorID, model.AuditUnpay, it.ID, paymentSnap(cur), nil))
	})
}

func paymentSnap(p model.ItemPayment) json.RawMessage {
	b, err := json.Marshal(struct {
		ID       string `json:"id"`
		ItemID   string `json:"item_id"`
		Date     string `json:"date"`
		Amount   int    `json:"amount"`
		Currency string `json:"currency"`
	}{ID: p.ID, ItemID: p.ItemID, Date: p.Date, Amount: p.Amount, Currency: p.Currency})
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

func paidDatesByItem(rows []model.ItemPayment) map[string]map[string]model.ItemPayment {
	out := map[string]map[string]model.ItemPayment{}
	for _, p := range rows {
		byDay, ok := out[p.ItemID]
		if !ok {
			byDay = map[string]model.ItemPayment{}
			out[p.ItemID] = byDay
		}
		byDay[p.Date] = p
	}
	return out
}

func paidDateSet(byDay map[string]model.ItemPayment) map[string]struct{} {
	out := map[string]struct{}{}
	for d := range byDay {
		out[d] = struct{}{}
	}
	return out
}

func nextOpenAt(it model.Item, today time.Time, paid map[string]struct{}) (string, bool, error) {
	if it.Status == model.StatusPaid {
		return "", false, nil
	}
	day, ok, err := nextUnpaidOccurrence(it, today, paid)
	if err != nil || !ok {
		return "", false, err
	}
	return day.Format(model.DateLayout), true, nil
}
