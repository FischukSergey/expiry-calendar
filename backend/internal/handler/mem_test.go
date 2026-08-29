package handler_test

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

type memKinds struct {
	mu   sync.Mutex
	byID map[string]model.Kind
}

func newMemKinds() *memKinds {
	return &memKinds{byID: map[string]model.Kind{}}
}

func (m *memKinds) List(context.Context) ([]model.Kind, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Kind, 0, len(m.byID))
	for _, k := range m.byID {
		out = append(out, k)
	}
	return out, nil
}

func (m *memKinds) ByID(_ context.Context, id string) (model.Kind, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k, ok := m.byID[id]
	if !ok {
		return model.Kind{}, model.ErrNotFound
	}
	return k, nil
}

func (m *memKinds) Create(_ context.Context, k model.Kind) (model.Kind, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k.ID = fixtureUUID
	m.byID[k.ID] = k
	return k, nil
}

func (m *memKinds) Update(_ context.Context, k model.Kind) (model.Kind, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byID[k.ID] = k
	return k, nil
}

func (m *memKinds) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.byID[id]; !ok {
		return model.ErrNotFound
	}
	delete(m.byID, id)
	return nil
}

func (m *memKinds) CountItems(context.Context, string) (int, error) { return 0, nil }
