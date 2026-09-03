package service_test

import (
	"context"
	"sync"
	"time"

	"duekeep/internal/model"
)

type memUsers struct {
	mu      sync.Mutex
	byEmail map[string]model.User
	byID    map[string]model.User
}

func newMemUsers() *memUsers {
	return &memUsers{byEmail: map[string]model.User{}, byID: map[string]model.User{}}
}

func (m *memUsers) Create(_ context.Context, email, passwordHash string, role model.Role) (model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byEmail[email]; ok {
		return model.User{}, model.ErrConflict
	}
	u := model.User{
		ID:           "user-" + email,
		Email:        email,
		Role:         role,
		PasswordHash: passwordHash,
		CreatedAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	m.byEmail[email] = u
	m.byID[u.ID] = u
	return u, nil
}

func (m *memUsers) ByEmail(_ context.Context, email string) (model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byEmail[email]
	if !ok {
		return model.User{}, model.ErrNotFound
	}
	return u, nil
}

func (m *memUsers) ByID(_ context.Context, id string) (model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byID[id]
	if !ok {
		return model.User{}, model.ErrNotFound
	}
	return u, nil
}

type memRefresh struct {
	mu     sync.Mutex
	byHash map[string]model.RefreshSession
}

func newMemRefresh() *memRefresh {
	return &memRefresh{byHash: map[string]model.RefreshSession{}}
}

func (m *memRefresh) Insert(_ context.Context, rec model.RefreshSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byHash[rec.TokenHash] = rec
	return nil
}

func (m *memRefresh) ByHash(_ context.Context, hash string) (model.RefreshSession, error) {
	return m.lookup(hash)
}

func (m *memRefresh) ByHashForUpdate(_ context.Context, hash string) (model.RefreshSession, error) {
	return m.lookup(hash)
}

func (m *memRefresh) lookup(hash string) (model.RefreshSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.byHash[hash]
	if !ok {
		return model.RefreshSession{}, model.ErrNotFound
	}
	return rec, nil
}

func (m *memRefresh) RevokeID(_ context.Context, id string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for h, rec := range m.byHash {
		if rec.ID == id && rec.RevokedAt == nil {
			rec.RevokedAt = &at
			m.byHash[h] = rec
		}
	}
	return nil
}

func (m *memRefresh) RevokeFamily(_ context.Context, familyID string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for h, rec := range m.byHash {
		if rec.FamilyID == familyID && rec.RevokedAt == nil {
			rec.RevokedAt = &at
			m.byHash[h] = rec
		}
	}
	return nil
}

func (m *memRefresh) RevokeUser(_ context.Context, userID string, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for h, rec := range m.byHash {
		if rec.UserID == userID && rec.RevokedAt == nil {
			rec.RevokedAt = &at
			m.byHash[h] = rec
		}
	}
	return nil
}

func nopTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type memPayments struct {
	mu   sync.Mutex
	rows []model.ItemPayment
}

func newMemPayments() *memPayments { return &memPayments{} }

func (m *memPayments) Insert(_ context.Context, p model.ItemPayment) (model.ItemPayment, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cur := range m.rows {
		if cur.ItemID == p.ItemID && cur.Date == p.Date {
			return cur, false, nil
		}
	}
	p.ID = "pay-" + p.ItemID + "-" + p.Date
	p.CreatedAt = time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	m.rows = append(m.rows, p)
	return p, true, nil
}

func (m *memPayments) GetByItemDate(_ context.Context, itemID, date string) (model.ItemPayment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cur := range m.rows {
		if cur.ItemID == itemID && cur.Date == date {
			return cur, nil
		}
	}
	return model.ItemPayment{}, model.ErrNotFound
}

func (m *memPayments) DeleteByItemDate(_ context.Context, itemID, date string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := make([]model.ItemPayment, 0, len(m.rows))
	for _, cur := range m.rows {
		if cur.ItemID == itemID && cur.Date == date {
			continue
		}
		kept = append(kept, cur)
	}
	m.rows = kept
	return nil
}

func (m *memPayments) ListByOwner(_ context.Context, ownerID string) ([]model.ItemPayment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.ItemPayment, 0)
	for _, cur := range m.rows {
		if cur.OwnerID == ownerID {
			out = append(out, cur)
		}
	}
	return out, nil
}

func (m *memPayments) ListByItemIDs(_ context.Context, itemIDs []string) ([]model.ItemPayment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	want := map[string]struct{}{}
	for _, id := range itemIDs {
		want[id] = struct{}{}
	}
	out := make([]model.ItemPayment, 0)
	for _, cur := range m.rows {
		if _, ok := want[cur.ItemID]; ok {
			out = append(out, cur)
		}
	}
	return out, nil
}
