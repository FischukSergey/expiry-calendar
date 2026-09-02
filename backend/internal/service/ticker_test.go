package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"duekeep/internal/clock"
	"duekeep/internal/model"
	"duekeep/internal/service"
)

const (
	itemTitleDomain = "Домен"
	expiresPast     = "2026-08-01"
	calendarDay     = "2026-08-21"
	otherOwner      = "other"
	expiresSep      = "2026-09-01"
	expiresSoon     = "2026-09-10"
)

type tickItems struct {
	mu   sync.Mutex
	byID map[string]model.Item
}

func newTickItems() *tickItems {
	return &tickItems{byID: map[string]model.Item{}}
}

func (m *tickItems) put(it model.Item) model.Item {
	if it.ID == "" {
		it.ID = uuid.NewString()
	}
	m.byID[it.ID] = it
	return it
}

func (m *tickItems) ListOpen(context.Context) ([]model.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.Item, 0)
	for _, it := range m.byID {
		if it.Status == model.StatusCancelled || it.Status == model.StatusArchived {
			continue
		}
		out = append(out, it)
	}
	return out, nil
}

func (m *tickItems) SetStatus(_ context.Context, id, status string) (model.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	it, ok := m.byID[id]
	if !ok || it.Status == model.StatusCancelled || it.Status == model.StatusArchived || it.Status == model.StatusPaid {
		return model.Item{}, model.ErrNotFound
	}
	it.Status = status
	m.byID[id] = it
	return it, nil
}

func (m *tickItems) get(id string) model.Item {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.byID[id]
}

type tickNotes struct {
	mu   sync.Mutex
	rows []model.Notification
}

func (m *tickNotes) Insert(_ context.Context, n model.Notification) (model.Notification, bool, error) {
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
	m.rows = append(m.rows, n)
	return n, true, nil
}

func (m *tickNotes) List(_ context.Context, ownerID string, unread bool, _ model.Page) ([]model.Notification, int, error) {
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
	return out, len(out), nil
}

func (m *tickNotes) MarkRead(context.Context, string, string) error { return nil }

func (m *tickNotes) MarkAllRead(context.Context, string) error { return nil }

func TestTickerMovesStatusAndNotifies(t *testing.T) {
	t.Parallel()
	store := newTickItems()
	notes := &tickNotes{}
	today := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	active := store.put(model.Item{
		OwnerID: "owner-1", Title: itemTitleDomain, Status: model.StatusActive,
		ExpiresAt: expiresSoon, NotifyBeforeDays: model.Ptr(30),
	})
	tkr := service.NewTicker(store, notes, nopTx, clock.Fixed{T: today}, nil)
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if got := store.get(active.ID).Status; got != model.StatusExpiring {
		t.Fatalf("status %s", got)
	}
	if len(notes.rows) != 1 || notes.rows[0].ToStatus != model.StatusExpiring || notes.rows[0].OwnerID != "owner-1" {
		t.Fatalf("notes %+v", notes.rows)
	}
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if len(notes.rows) != 1 {
		t.Fatalf("duplicate notes %d", len(notes.rows))
	}
}

func TestTickerSkipsCancelledArchived(t *testing.T) {
	t.Parallel()
	store := newTickItems()
	notes := &tickNotes{}
	cancelled := store.put(model.Item{
		Title: "Отмена", Status: model.StatusCancelled,
		ExpiresAt: expiresPast, NotifyBeforeDays: model.Ptr(30),
	})
	archived := store.put(model.Item{
		Title: "Архив", Status: model.StatusArchived,
		ExpiresAt: expiresPast, NotifyBeforeDays: model.Ptr(30),
	})
	tkr := service.NewTicker(store, notes, nopTx, clock.Fixed{
		T: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC),
	}, nil)
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if store.get(cancelled.ID).Status != model.StatusCancelled {
		t.Fatal("cancelled changed")
	}
	if store.get(archived.ID).Status != model.StatusArchived {
		t.Fatal("archived changed")
	}
	if len(notes.rows) != 0 {
		t.Fatalf("notes %+v", notes.rows)
	}
}

func TestTickerExpiresAndUniqueDay(t *testing.T) {
	t.Parallel()
	store := newTickItems()
	notes := &tickNotes{}
	it := store.put(model.Item{
		Title: "Полис", Status: model.StatusExpiring,
		ExpiresAt: "2026-08-25", NotifyBeforeDays: model.Ptr(30),
	})
	tkr := service.NewTicker(store, notes, nopTx, clock.Fixed{
		T: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC),
	}, nil)
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if store.get(it.ID).Status != model.StatusExpired {
		t.Fatalf("status %s", store.get(it.ID).Status)
	}
	if len(notes.rows) != 1 || notes.rows[0].ToStatus != model.StatusExpired {
		t.Fatalf("notes %+v", notes.rows)
	}
}

func TestTickerRunDisabled(t *testing.T) {
	t.Parallel()
	tkr := service.NewTicker(newTickItems(), &tickNotes{}, nopTx, clock.Fixed{}, nil)
	tkr.Run(t.Context(), 0)
}

type recBus struct {
	mu  sync.Mutex
	got []model.Notification
}

func (r *recBus) Notify(n model.Notification) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.got = append(r.got, n)
}

func TestTickerPublishesToBus(t *testing.T) {
	t.Parallel()
	store := newTickItems()
	notes := &tickNotes{}
	bus := &recBus{}
	store.put(model.Item{
		Title: itemTitleDomain, Status: model.StatusActive,
		ExpiresAt: expiresSoon, NotifyBeforeDays: model.Ptr(30),
	})
	tkr := service.NewTicker(store, notes, nopTx, clock.Fixed{
		T: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC),
	}, bus)
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	bus.mu.Lock()
	n := len(bus.got)
	bus.mu.Unlock()
	if n != 1 {
		t.Fatalf("bus %d", n)
	}
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	bus.mu.Lock()
	n = len(bus.got)
	bus.mu.Unlock()
	if n != 1 {
		t.Fatalf("duplicate bus %d", n)
	}
}

func TestTickerSkipsPaid(t *testing.T) {
	t.Parallel()
	store := newTickItems()
	notes := &tickNotes{}
	paid := store.put(model.Item{
		Title: itemTitleDomain, Status: model.StatusPaid,
		ExpiresAt: expiresSoon, NotifyBeforeDays: model.Ptr(30),
	})
	tkr := service.NewTicker(store, notes, nopTx, clock.Fixed{
		T: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC),
	}, nil)
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if store.get(paid.ID).Status != model.StatusPaid {
		t.Fatal("paid changed")
	}
	if len(notes.rows) != 0 {
		t.Fatalf("notes %+v", notes.rows)
	}
}

func TestTickerNullNotifyNoExpiring(t *testing.T) {
	t.Parallel()
	store := newTickItems()
	notes := &tickNotes{}
	it := store.put(model.Item{
		Title: itemTitleDomain, Status: model.StatusActive,
		ExpiresAt: expiresSoon, NotifyBeforeDays: nil,
	})
	tkr := service.NewTicker(store, notes, nopTx, clock.Fixed{
		T: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC),
	}, nil)
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if store.get(it.ID).Status != model.StatusActive {
		t.Fatalf("status %s", store.get(it.ID).Status)
	}
	if len(notes.rows) != 0 {
		t.Fatalf("notes %+v", notes.rows)
	}
}

func TestTickerNullNotifyExpiresWithoutNote(t *testing.T) {
	t.Parallel()
	store := newTickItems()
	notes := &tickNotes{}
	it := store.put(model.Item{
		Title: "Полис", Status: model.StatusActive,
		ExpiresAt: expiresPast, NotifyBeforeDays: nil,
	})
	tkr := service.NewTicker(store, notes, nopTx, clock.Fixed{
		T: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC),
	}, nil)
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if store.get(it.ID).Status != model.StatusExpired {
		t.Fatalf("status %s", store.get(it.ID).Status)
	}
	if len(notes.rows) != 0 {
		t.Fatalf("notes %+v", notes.rows)
	}
}
