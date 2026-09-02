package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"duekeep/internal/clock"
	"duekeep/internal/model"
)

const (
	fieldTitle      = "title"
	fieldKindID     = "kind_id"
	fieldExpiresAt  = "expires_at"
	fieldCost       = "cost_amount"
	fieldCurrency   = "currency"
	fieldBilling    = "billing_period"
	fieldStartedAt  = "started_at"
	fieldStatus     = "status"
	fieldCategory   = "category_id"
	fieldSort       = "sort"
	fieldOrder      = "order"
	fieldIDs        = "ids"
	fieldNewExpires = "new_expires_at"
	fieldNewCost    = "new_cost"
	fieldNotify     = "notify_before_days"
)

// ItemStore — SQL к items.
type ItemStore interface {
	Create(ctx context.Context, it model.Item) (model.Item, error)
	ByID(ctx context.Context, id string) (model.Item, error)
	Update(ctx context.Context, it model.Item) (model.Item, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, f model.ItemFilter, page model.Page) ([]model.Item, int, error)
	BulkUpdate(ctx context.Context, ids []string, categoryID *string, status *string, ownerID string) (int, error)
}

// RenewalStore — история продлений.
type RenewalStore interface {
	Create(ctx context.Context, r model.Renewal) (model.Renewal, error)
	ListByItem(ctx context.Context, itemID string) ([]model.Renewal, error)
}

// AuditStore — журнал.
type AuditStore interface {
	Create(ctx context.Context, e model.AuditEntry) error
	List(ctx context.Context, ownerID string, page model.Page) ([]model.AuditEntry, int, error)
}

// Item — CRUD записей, renew, bulk, audit.
type Item struct {
	items    ItemStore
	kinds    KindStore
	cats     CategoryStore
	renewals RenewalStore
	audit    AuditStore
	notes    NotificationStore
	bus      EventBus
	tx       TxFunc
	clk      clock.Clock
}

// NewItem собирает сервис записей.
func NewItem(
	items ItemStore,
	kinds KindStore,
	cats CategoryStore,
	renewals RenewalStore,
	audit AuditStore,
	tx TxFunc,
	clk clock.Clock,
) *Item {
	return &Item{items: items, kinds: kinds, cats: cats, renewals: renewals, audit: audit, tx: tx, clk: clk}
}

// SetNotify включает ленту и SSE/push при переходе в expiring/expired на записи.
// Без вызова тикер остаётся единственным источником (тесты).
func (s *Item) SetNotify(notes NotificationStore, bus EventBus) {
	s.notes = notes
	s.bus = bus
}

// List фильтрует и пагинирует. Только свой owner_id; category_id включает потомков.
func (s *Item) List(ctx context.Context, f model.ItemFilter, page model.Page, actorID string) (model.ItemList, error) {
	f, err := s.scopedFilter(ctx, f, actorID)
	if err != nil {
		return model.ItemList{}, err
	}
	rows, total, err := s.items.List(ctx, f, page)
	if err != nil {
		return model.ItemList{}, err
	}
	if rows == nil {
		rows = []model.Item{}
	}
	return model.ItemList{Items: rows, Page: page.Page, PerPage: page.PerPage, Total: total}, nil
}

// Create валидирует attrs/даты и считает status.
func (s *Item) Create(ctx context.Context, in model.Item, actorID string) (model.Item, error) {
	it, err := s.prepareWrite(ctx, in, actorID)
	if err != nil {
		return model.Item{}, err
	}
	it.OwnerID = actorID
	var created model.Item
	var note model.Notification
	var inserted bool
	err = s.tx(ctx, func(ctx context.Context) error {
		var cerr error
		created, cerr = s.items.Create(ctx, it)
		if cerr != nil {
			return cerr
		}
		if err := s.audit.Create(ctx, auditEntry(actorID, model.AuditCreate, created.ID, nil, itemSnap(created))); err != nil {
			return err
		}
		note, inserted, cerr = s.notifyTransition(ctx, "", created)
		return cerr
	})
	if err == nil {
		s.publishNote(note, inserted)
	}
	return created, err
}

// Get карточка + renewals. Чужой id → 404.
func (s *Item) Get(ctx context.Context, id, actorID string) (model.ItemCard, error) {
	it, err := s.items.ByID(ctx, id)
	if err != nil {
		return model.ItemCard{}, err
	}
	if err := requireOwner(it.OwnerID, actorID); err != nil {
		return model.ItemCard{}, err
	}
	hist, err := s.renewals.ListByItem(ctx, id)
	if err != nil {
		return model.ItemCard{}, err
	}
	if hist == nil {
		hist = []model.Renewal{}
	}
	return model.ItemCard{Item: it, Renewals: hist}, nil
}

// Patch частичное обновление. Пустой patch не ошибка.
func (s *Item) Patch(ctx context.Context, id string, p model.ItemPatch, actorID string) (model.Item, error) {
	cur, err := s.items.ByID(ctx, id)
	if err != nil {
		return model.Item{}, err
	}
	if err := requireOwner(cur.OwnerID, actorID); err != nil {
		return model.Item{}, err
	}
	before := itemSnap(cur)
	prevStatus := cur.Status
	applyItemPatch(&cur, p)
	it, err := s.prepareWrite(ctx, cur, actorID)
	if err != nil {
		return model.Item{}, err
	}
	it.ID = id
	var updated model.Item
	var note model.Notification
	var inserted bool
	err = s.tx(ctx, func(ctx context.Context) error {
		var uerr error
		updated, uerr = s.items.Update(ctx, it)
		if uerr != nil {
			return uerr
		}
		if err := s.audit.Create(ctx, auditEntry(actorID, model.AuditUpdate, id, before, itemSnap(updated))); err != nil {
			return err
		}
		note, inserted, uerr = s.notifyTransition(ctx, prevStatus, updated)
		return uerr
	})
	if err == nil {
		s.publishNote(note, inserted)
	}
	return updated, err
}

// Delete пишет audit и удаляет строку.
func (s *Item) Delete(ctx context.Context, id, actorID string) error {
	cur, err := s.items.ByID(ctx, id)
	if err != nil {
		return err
	}
	if err := requireOwner(cur.OwnerID, actorID); err != nil {
		return err
	}
	before := itemSnap(cur)
	return s.tx(ctx, func(ctx context.Context) error {
		if err := s.items.Delete(ctx, id); err != nil {
			return err
		}
		return s.audit.Create(ctx, auditEntry(actorID, model.AuditDelete, id, before, nil))
	})
}

// Renew сдвигает срок, пишет renewals и audit.
func (s *Item) Renew(ctx context.Context, id string, in model.RenewInput, actorID string) (model.Item, error) {
	cur, err := s.items.ByID(ctx, id)
	if err != nil {
		return model.Item{}, err
	}
	if err := requireOwner(cur.OwnerID, actorID); err != nil {
		return model.Item{}, err
	}
	expires, err := parseDate(fieldNewExpires, in.NewExpiresAt)
	if err != nil {
		return model.Item{}, err
	}
	newCost := cur.CostAmount
	if in.NewCost != nil {
		if *in.NewCost < 0 {
			return model.Item{}, model.Validation("invalid cost", map[string]any{fieldNewCost: detailMinZero})
		}
		newCost = *in.NewCost
	}
	before := itemSnap(cur)
	oldExp := cur.ExpiresAt
	oldCost := cur.CostAmount
	cur.ExpiresAt = expires.Format(model.DateLayout)
	cur.CostAmount = newCost
	if cur.Status == model.StatusPaid {
		cur.Status = ""
	}
	it, err := s.prepareWrite(ctx, cur, actorID)
	if err != nil {
		return model.Item{}, err
	}
	it.ID = id
	rec := model.Renewal{
		ItemID:       id,
		ActorID:      actorID,
		OldExpiresAt: oldExp,
		NewExpiresAt: it.ExpiresAt,
		OldCost:      oldCost,
		NewCost:      newCost,
		Comment:      strings.TrimSpace(in.Comment),
	}
	var updated model.Item
	err = s.tx(ctx, func(ctx context.Context) error {
		if _, err := s.renewals.Create(ctx, rec); err != nil {
			return err
		}
		var uerr error
		updated, uerr = s.items.Update(ctx, it)
		if uerr != nil {
			return uerr
		}
		return s.audit.Create(ctx, auditEntry(actorID, model.AuditRenew, id, before, itemSnap(updated)))
	})
	return updated, err
}

// Bulk меняет category_id и/или status (только cancelled/archived/paid).
func (s *Item) Bulk(ctx context.Context, in model.BulkInput, actorID string) (model.BulkResult, error) {
	if len(in.IDs) == 0 {
		return model.BulkResult{}, model.Validation("ids required", map[string]any{fieldIDs: detailRequired})
	}
	if in.CategoryID == nil && in.Status == nil {
		return model.BulkResult{}, model.Validation("nothing to update", map[string]any{
			fieldCategory: detailRequired, fieldStatus: detailRequired,
		})
	}
	for _, id := range in.IDs {
		if _, err := uuid.Parse(id); err != nil {
			return model.BulkResult{}, model.Validation("invalid id", map[string]any{fieldIDs: id})
		}
		it, err := s.items.ByID(ctx, id)
		if err != nil {
			return model.BulkResult{}, err
		}
		if err := requireOwner(it.OwnerID, actorID); err != nil {
			return model.BulkResult{}, err
		}
	}
	if in.CategoryID != nil {
		if err := s.ownCategory(ctx, *in.CategoryID, actorID); err != nil {
			return model.BulkResult{}, err
		}
	}
	if in.Status != nil {
		switch *in.Status {
		case model.StatusCancelled, model.StatusArchived, model.StatusPaid:
		default:
			return model.BulkResult{}, model.Validation("invalid status", map[string]any{fieldStatus: "cancelled|archived|paid"})
		}
	}
	var updated int
	err := s.tx(ctx, func(ctx context.Context) error {
		n, err := s.items.BulkUpdate(ctx, in.IDs, in.CategoryID, in.Status, actorID)
		if err != nil {
			return err
		}
		updated = n
		after, merr := json.Marshal(struct {
			IDs        []string `json:"ids"`
			CategoryID *string  `json:"category_id"`
			Status     *string  `json:"status"`
			Updated    int      `json:"updated"`
		}{IDs: in.IDs, CategoryID: in.CategoryID, Status: in.Status, Updated: n})
		if merr != nil {
			return merr
		}
		entityID := in.IDs[0]
		return s.audit.Create(ctx, auditEntry(actorID, model.AuditBulk, entityID, nil, after))
	})
	return model.BulkResult{Updated: updated}, err
}

// ListAudit — GET /audit, только события владельца.
func (s *Item) ListAudit(ctx context.Context, page model.Page, actorID string) (model.AuditList, error) {
	rows, total, err := s.audit.List(ctx, actorID, page)
	if err != nil {
		return model.AuditList{}, err
	}
	if rows == nil {
		rows = []model.AuditEntry{}
	}
	return model.AuditList{Items: rows, Page: page.Page, PerPage: page.PerPage, Total: total}, nil
}

func (s *Item) prepareWrite(ctx context.Context, in model.Item, actorID string) (model.Item, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.Vendor = strings.TrimSpace(in.Vendor)
	in.URL = strings.TrimSpace(in.URL)
	in.AccountHint = strings.TrimSpace(in.AccountHint)
	in.KindID = strings.TrimSpace(in.KindID)
	in.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	in.BillingPeriod = strings.TrimSpace(in.BillingPeriod)
	if in.Tags == nil {
		in.Tags = []string{}
	}
	if in.Attrs == nil {
		in.Attrs = map[string]any{}
	}
	if in.Title == "" {
		return model.Item{}, model.Validation("invalid title", map[string]any{fieldTitle: detailRequired})
	}
	if _, err := uuid.Parse(in.KindID); err != nil {
		return model.Item{}, model.Validation("invalid kind_id", map[string]any{fieldKindID: detailUUID})
	}
	kind, err := s.kinds.ByID(ctx, in.KindID)
	if err != nil {
		return model.Item{}, err
	}
	if err := ValidateAttrs(kind.AttrSchema, in.Attrs); err != nil {
		return model.Item{}, err
	}
	if err := s.normalizeCategoryID(ctx, &in, actorID); err != nil {
		return model.Item{}, err
	}
	if in.CostAmount < 0 {
		return model.Item{}, model.Validation("invalid cost", map[string]any{fieldCost: detailMinZero})
	}
	if in.Currency == "" {
		in.Currency = model.CurrencyRUB
	}
	if len(in.Currency) != 3 {
		return model.Item{}, model.Validation("invalid currency", map[string]any{fieldCurrency: "ISO 4217"})
	}
	if in.BillingPeriod == "" {
		in.BillingPeriod = model.BillingOneTime
	}
	switch in.BillingPeriod {
	case model.BillingOneTime, model.BillingMonthly, model.BillingYearly:
	default:
		return model.Item{}, model.Validation("invalid billing_period", map[string]any{fieldBilling: "one_time|monthly|yearly"})
	}
	if in.NotifyBeforeDays != nil && *in.NotifyBeforeDays < 0 {
		return model.Item{}, model.Validation("invalid notify_before_days", map[string]any{fieldNotify: detailMinZero})
	}
	expires, err := parseDate(fieldExpiresAt, in.ExpiresAt)
	if err != nil {
		return model.Item{}, err
	}
	if in.StartedAt != nil {
		started, err := parseDate(fieldStartedAt, *in.StartedAt)
		if err != nil {
			return model.Item{}, err
		}
		if started.After(expires) {
			return model.Item{}, model.Validation("invalid dates", map[string]any{fieldStartedAt: "<= expires_at"})
		}
		s := started.Format(model.DateLayout)
		in.StartedAt = &s
	}
	in.ExpiresAt = expires.Format(model.DateLayout)
	in.Status = StatusAtWrite(clock.Today(s.clk), expires, in.NotifyBeforeDays, in.Status)
	return in, nil
}

func (s *Item) scopedFilter(ctx context.Context, f model.ItemFilter, actorID string) (model.ItemFilter, error) {
	f, err := normalizeFilter(f)
	if err != nil {
		return f, err
	}
	f.OwnerID = actorID
	if f.CategoryID != "" {
		if err := s.ownCategory(ctx, f.CategoryID, actorID); err != nil {
			return f, err
		}
		ids, err := s.cats.DescendantIDs(ctx, f.CategoryID)
		if err != nil {
			return f, err
		}
		f.CategoryIDs = ids
	}
	return f, nil
}

func (s *Item) ownCategory(ctx context.Context, id, actorID string) error {
	cat, err := s.cats.ByID(ctx, id)
	if err != nil {
		return err
	}
	return requireOwner(cat.OwnerID, actorID)
}

func (s *Item) normalizeCategoryID(ctx context.Context, in *model.Item, actorID string) error {
	if in.CategoryID == nil {
		return nil
	}
	cid := strings.TrimSpace(*in.CategoryID)
	if cid == "" {
		in.CategoryID = nil
		return nil
	}
	if _, err := uuid.Parse(cid); err != nil {
		return model.Validation("invalid category_id", map[string]any{fieldCategory: detailUUID})
	}
	if err := s.ownCategory(ctx, cid, actorID); err != nil {
		return err
	}
	in.CategoryID = &cid
	return nil
}

func applyItemPatch(cur *model.Item, p model.ItemPatch) {
	if p.Title != nil {
		cur.Title = *p.Title
	}
	if p.Description != nil {
		cur.Description = *p.Description
	}
	if p.KindID != nil {
		cur.KindID = *p.KindID
	}
	if p.SetCategory {
		cur.CategoryID = p.CategoryID
	}
	if p.Vendor != nil {
		cur.Vendor = *p.Vendor
	}
	if p.Tags != nil {
		cur.Tags = *p.Tags
	}
	if p.CostAmount != nil {
		cur.CostAmount = *p.CostAmount
	}
	if p.Currency != nil {
		cur.Currency = *p.Currency
	}
	if p.BillingPeriod != nil {
		cur.BillingPeriod = *p.BillingPeriod
	}
	if p.SetStarted {
		cur.StartedAt = p.StartedAt
	}
	if p.ExpiresAt != nil {
		cur.ExpiresAt = *p.ExpiresAt
	}
	if p.SetNotify {
		cur.NotifyBeforeDays = p.NotifyBeforeDays
	}
	if p.URL != nil {
		cur.URL = *p.URL
	}
	if p.AccountHint != nil {
		cur.AccountHint = *p.AccountHint
	}
	if p.Status != nil {
		cur.Status = *p.Status
	}
	if p.SetAttrs {
		cur.Attrs = p.Attrs
	}
}

func normalizeFilter(f model.ItemFilter) (model.ItemFilter, error) {
	f.Q = strings.TrimSpace(f.Q)
	f.KindID = strings.TrimSpace(f.KindID)
	f.Status = strings.TrimSpace(f.Status)
	f.CategoryID = strings.TrimSpace(f.CategoryID)
	f.Vendor = strings.TrimSpace(f.Vendor)
	f.BillingPeriod = strings.TrimSpace(f.BillingPeriod)
	f.Tag = strings.TrimSpace(f.Tag)
	f.Sort = strings.TrimSpace(f.Sort)
	f.Order = strings.ToLower(strings.TrimSpace(f.Order))
	if f.KindID != "" {
		if _, err := uuid.Parse(f.KindID); err != nil {
			return f, model.Validation("invalid kind_id", map[string]any{fieldKindID: detailUUID})
		}
	}
	if f.CategoryID != "" {
		if _, err := uuid.Parse(f.CategoryID); err != nil {
			return f, model.Validation("invalid category_id", map[string]any{fieldCategory: detailUUID})
		}
	}
	if f.Status != "" {
		switch f.Status {
		case model.StatusActive, model.StatusExpiring, model.StatusExpired,
			model.StatusCancelled, model.StatusArchived, model.StatusPaid:
		default:
			return f, model.Validation("invalid status", map[string]any{fieldStatus: "enum"})
		}
	}
	if f.BillingPeriod != "" {
		switch f.BillingPeriod {
		case model.BillingOneTime, model.BillingMonthly, model.BillingYearly:
		default:
			return f, model.Validation("invalid billing_period", map[string]any{fieldBilling: "enum"})
		}
	}
	if f.ExpiresFrom != "" {
		if _, err := parseDate("expires_from", f.ExpiresFrom); err != nil {
			return f, err
		}
	}
	if f.ExpiresTo != "" {
		if _, err := parseDate("expires_to", f.ExpiresTo); err != nil {
			return f, err
		}
	}
	if f.Sort == "" {
		f.Sort = model.SortExpiresAt
	}
	switch f.Sort {
	case model.SortExpiresAt, model.SortCostAmount, model.SortTitle, model.SortUpdatedAt:
	default:
		return f, model.Validation("invalid sort", map[string]any{fieldSort: "expires_at|cost_amount|title|updated_at"})
	}
	if f.Order == "" {
		f.Order = model.OrderAsc
	}
	if f.Order != model.OrderAsc && f.Order != model.OrderDesc {
		return f, model.Validation("invalid order", map[string]any{fieldOrder: "asc|desc"})
	}
	return f, nil
}

// notifyTransition пишет unread, если статус стал expiring/expired. Повтор за день — false.
func (s *Item) notifyTransition(ctx context.Context, prev string, it model.Item) (model.Notification, bool, error) {
	if s.notes == nil {
		return model.Notification{}, false, nil
	}
	if it.NotifyBeforeDays == nil {
		return model.Notification{}, false, nil
	}
	if it.Status != model.StatusExpiring && it.Status != model.StatusExpired {
		return model.Notification{}, false, nil
	}
	if prev == it.Status {
		return model.Notification{}, false, nil
	}
	return s.notes.Insert(ctx, model.Notification{
		OwnerID:   it.OwnerID,
		ItemID:    it.ID,
		ToStatus:  it.Status,
		Title:     it.Title,
		CreatedAt: s.clk.Now().UTC(),
	})
}

func (s *Item) publishNote(n model.Notification, inserted bool) {
	if inserted && s.bus != nil {
		s.bus.Notify(n)
	}
}

func parseDate(field, raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, model.Validation("invalid date", map[string]any{field: detailRequired})
	}
	t, err := time.Parse(model.DateLayout, raw)
	if err != nil {
		return time.Time{}, model.Validation("invalid date", map[string]any{field: model.DateLayout})
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
}
