package handler_test

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"duekeep/internal/model"
)

type memItems struct {
	mu   sync.Mutex
	byID map[string]model.Item
}

func newMemItems() *memItems {
	return &memItems{byID: map[string]model.Item{}}
}

func (m *memItems) Create(_ context.Context, it model.Item) (model.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	it.ID = uuid.NewString()
	now := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	it.CreatedAt = now
	it.UpdatedAt = now
	if it.Tags == nil {
		it.Tags = []string{}
	}
	if it.Attrs == nil {
		it.Attrs = map[string]any{}
	}
	m.byID[it.ID] = it
	return it, nil
}

func (m *memItems) ByID(_ context.Context, id string) (model.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.byID[id]
	if !ok {
		return model.Item{}, model.ErrNotFound
	}
	return it, nil
}

func (m *memItems) Update(_ context.Context, it model.Item) (model.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.byID[it.ID]
	if !ok {
		return model.Item{}, model.ErrNotFound
	}
	it.CreatedAt = cur.CreatedAt
	it.UpdatedAt = time.Date(2026, 8, 26, 13, 0, 0, 0, time.UTC)
	m.byID[it.ID] = it
	return it, nil
}

func (m *memItems) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byID[id]; !ok {
		return model.ErrNotFound
	}
	delete(m.byID, id)
	return nil
}

func (m *memItems) List(_ context.Context, f model.ItemFilter, page model.Page) ([]model.Item, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Item, 0)
	for _, it := range m.byID {
		if !memItemMatch(it, f) {
			continue
		}
		out = append(out, it)
	}
	slices.SortFunc(out, func(a, b model.Item) int {
		return strings.Compare(a.ID, b.ID)
	})
	total := len(out)
	start := min(page.Offset(), total)
	end := min(start+page.PerPage, total)
	return out[start:end], total, nil
}

func (m *memItems) ListOpen(_ context.Context) ([]model.Item, error) {
	return m.listOpen("")
}

func (m *memItems) ListOpenByOwner(_ context.Context, ownerID string) ([]model.Item, error) {
	return m.listOpen(ownerID)
}

func (m *memItems) listOpen(ownerID string) ([]model.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Item, 0)
	for _, it := range m.byID {
		if it.Status == model.StatusCancelled || it.Status == model.StatusArchived {
			continue
		}
		if ownerID != "" && it.OwnerID != ownerID {
			continue
		}
		out = append(out, it)
	}
	slices.SortFunc(out, func(a, b model.Item) int {
		return strings.Compare(a.ID, b.ID)
	})
	return out, nil
}

func (m *memItems) SetStatus(_ context.Context, id, status string) (model.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.byID[id]
	if !ok || it.Status == model.StatusCancelled || it.Status == model.StatusArchived {
		return model.Item{}, model.ErrNotFound
	}
	it.Status = status
	it.UpdatedAt = time.Date(2026, 8, 26, 13, 0, 0, 0, time.UTC)
	m.byID[id] = it
	return it, nil
}

func (m *memItems) BulkUpdate(_ context.Context, ids []string, categoryID *string, status *string, ownerID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, id := range ids {
		it, ok := m.byID[id]
		if !ok || it.OwnerID != ownerID {
			continue
		}
		if categoryID != nil {
			it.CategoryID = categoryID
		}
		if status != nil {
			it.Status = *status
		}
		m.byID[id] = it
		n++
	}
	return n, nil
}

func memItemMatch(it model.Item, f model.ItemFilter) bool {
	if it.OwnerID != f.OwnerID {
		return false
	}
	if f.KindID != "" && it.KindID != f.KindID {
		return false
	}
	if f.Status != "" && it.Status != f.Status {
		return false
	}
	if len(f.CategoryIDs) > 0 {
		if it.CategoryID == nil || !slices.Contains(f.CategoryIDs, *it.CategoryID) {
			return false
		}
	}
	if f.Vendor != "" && !strings.EqualFold(it.Vendor, f.Vendor) {
		return false
	}
	if f.BillingPeriod != "" && it.BillingPeriod != f.BillingPeriod {
		return false
	}
	if f.Tag != "" && !slices.Contains(it.Tags, f.Tag) {
		return false
	}
	if f.Q != "" {
		q := strings.ToLower(f.Q)
		ok := strings.Contains(strings.ToLower(it.Title), q) || strings.Contains(strings.ToLower(it.Vendor), q)
		for _, tag := range it.Tags {
			if strings.Contains(strings.ToLower(tag), q) {
				ok = true
			}
		}
		if !ok {
			return false
		}
	}
	if f.ExpiresFrom != "" && it.ExpiresAt < f.ExpiresFrom {
		return false
	}
	if f.ExpiresTo != "" && it.ExpiresAt > f.ExpiresTo {
		return false
	}
	if f.CostFrom != nil && it.CostAmount < *f.CostFrom {
		return false
	}
	if f.CostTo != nil && it.CostAmount > *f.CostTo {
		return false
	}
	return true
}

type memRenewals struct {
	mu   sync.Mutex
	rows []model.Renewal
}

func newMemRenewals() *memRenewals { return &memRenewals{} }

func (m *memRenewals) Create(_ context.Context, r model.Renewal) (model.Renewal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r.ID = uuid.NewString()
	r.CreatedAt = time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	m.rows = append(m.rows, r)
	return r, nil
}

func (m *memRenewals) ListByItem(_ context.Context, itemID string) ([]model.Renewal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Renewal, 0)
	for _, r := range m.rows {
		if r.ItemID == itemID {
			out = append(out, r)
		}
	}
	return out, nil
}

type memAudit struct {
	mu   sync.Mutex
	rows []model.AuditEntry
}

func newMemAudit() *memAudit { return &memAudit{} }

func (m *memAudit) Create(_ context.Context, e model.AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e.ID = uuid.NewString()
	e.CreatedAt = time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	m.rows = append(m.rows, e)
	return nil
}

func (m *memAudit) List(_ context.Context, ownerID string, page model.Page) ([]model.AuditEntry, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	owned := make([]model.AuditEntry, 0, len(m.rows))
	for _, e := range m.rows {
		if e.OwnerID == ownerID {
			owned = append(owned, e)
		}
	}
	total := len(owned)
	start := min(page.Offset(), total)
	end := min(start+page.PerPage, total)
	out := append([]model.AuditEntry(nil), owned[start:end]...)
	return out, total, nil
}

type memCats struct {
	mu   sync.Mutex
	byID map[string]model.Category
}

func newMemCats() *memCats {
	return &memCats{byID: map[string]model.Category{}}
}

func (m *memCats) List(_ context.Context, ownerID string) ([]model.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Category, 0, len(m.byID))
	for _, c := range m.byID {
		if c.OwnerID == ownerID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *memCats) ByID(_ context.Context, id string) (model.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.byID[id]
	if !ok {
		return model.Category{}, model.ErrNotFound
	}
	return c, nil
}

func (m *memCats) Create(_ context.Context, c model.Category) (model.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.ID = uuid.NewString()
	c.Children = []model.Category{}
	m.byID[c.ID] = c
	return c, nil
}

func (m *memCats) Update(_ context.Context, c model.Category) (model.Category, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byID[c.ID] = c
	return c, nil
}

func (m *memCats) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.byID, id)
	return nil
}

func (m *memCats) CountChildren(context.Context, string) (int, error) { return 0, nil }

func (m *memCats) CountItems(context.Context, string) (int, error) { return 0, nil }

func (m *memCats) DescendantIDs(_ context.Context, id string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []string{id}
	var walk func(string)
	walk = func(parent string) {
		for _, c := range m.byID {
			if c.ParentID != nil && *c.ParentID == parent {
				out = append(out, c.ID)
				walk(c.ID)
			}
		}
	}
	walk(id)
	return out, nil
}
