package handler_test

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"

	"duekeep/internal/model"
)

type memNotifications struct {
	mu   sync.Mutex
	rows []model.Notification
}

func newMemNotifications() *memNotifications {
	return &memNotifications{}
}

func (m *memNotifications) Insert(_ context.Context, n model.Notification) (model.Notification, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	day := n.CreatedAt.UTC().Format(model.DateLayout)
	for _, cur := range m.rows {
		if cur.ItemID == n.ItemID && cur.ToStatus == n.ToStatus &&
			cur.CreatedAt.UTC().Format(model.DateLayout) == day {
			return model.Notification{}, false, nil
		}
	}
	n.ID = uuid.NewString()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	}
	m.rows = append(m.rows, n)
	return n, true, nil
}

func (m *memNotifications) List(
	_ context.Context, ownerID string, unread bool, page model.Page,
) ([]model.Notification, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Notification, 0)
	for _, n := range m.rows {
		if n.OwnerID != ownerID {
			continue
		}
		if unread && n.ReadAt != nil {
			continue
		}
		out = append(out, n)
	}
	slices.SortFunc(out, func(a, b model.Notification) int {
		if a.CreatedAt.After(b.CreatedAt) {
			return -1
		}
		if a.CreatedAt.Before(b.CreatedAt) {
			return 1
		}
		return 0
	})
	total := len(out)
	start := min(page.Offset(), total)
	end := min(start+page.PerPage, total)
	return out[start:end], total, nil
}

func (m *memNotifications) MarkRead(_ context.Context, id, ownerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, n := range m.rows {
		if n.ID != id || n.OwnerID != ownerID {
			continue
		}
		if n.ReadAt == nil {
			now := time.Date(2026, 8, 26, 14, 0, 0, 0, time.UTC)
			n.ReadAt = &now
			m.rows[i] = n
		}
		return nil
	}
	return model.ErrNotFound
}

func (m *memNotifications) MarkAllRead(_ context.Context, ownerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Date(2026, 8, 26, 14, 0, 0, 0, time.UTC)
	for i, n := range m.rows {
		if n.OwnerID == ownerID && n.ReadAt == nil {
			n.ReadAt = &now
			m.rows[i] = n
		}
	}
	return nil
}
