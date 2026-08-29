package service_test

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"duekeep/internal/model"
	"duekeep/internal/service"
)

type memPushStore struct {
	mu   sync.Mutex
	byEP map[string]model.PushSubscription
}

func newMemPushStore() *memPushStore {
	return &memPushStore{byEP: map[string]model.PushSubscription{}}
}

func (m *memPushStore) Upsert(_ context.Context, s model.PushSubscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byEP[s.Endpoint] = s
	return nil
}

func (m *memPushStore) DeleteByEndpoint(_ context.Context, endpoint string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.byEP, endpoint)
	return nil
}

func (m *memPushStore) List(context.Context) ([]model.PushSubscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.PushSubscription, 0, len(m.byEP))
	for _, s := range m.byEP {
		out = append(out, s)
	}
	return out, nil
}

type stubSender struct {
	mu     sync.Mutex
	n      int
	status int
}

func (s *stubSender) Send(context.Context, model.PushSubscription, []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.n++
	if s.status == 0 {
		return 201, nil
	}
	return s.status, nil
}

func TestPushSubscribeValidates(t *testing.T) {
	t.Parallel()
	p := service.NewPush(newMemPushStore(), nil, "pub")
	err := p.Subscribe(t.Context(), "u1", model.PushSubscribe{}, "")
	if err == nil {
		t.Fatal("expected validation")
	}
	err = p.Subscribe(t.Context(), "u1", model.PushSubscribe{
		Endpoint: "https://push.example/a",
		Keys:     model.PushKeys{P256dh: "p", Auth: "a"},
	}, "ua")
	if err != nil {
		t.Fatal(err)
	}
}

func TestPushBroadcastDeletesOn410(t *testing.T) {
	t.Parallel()
	store := newMemPushStore()
	sender := &stubSender{status: http.StatusGone}
	p := service.NewPush(store, sender, "pub")
	if err := store.Upsert(t.Context(), model.PushSubscription{
		UserID: "u1", Endpoint: "https://push.example/gone", P256dh: "p", Auth: "a",
	}); err != nil {
		t.Fatal(err)
	}
	if err := p.Broadcast(t.Context(), model.Notification{
		ID: "n1", ItemID: "i1", ToStatus: model.StatusExpiring, Title: itemTitleDomain,
		CreatedAt: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}
	subs, err := store.List(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 0 {
		t.Fatalf("left %d", len(subs))
	}
}

func TestFanoutNotifiesSSEAndPush(t *testing.T) {
	t.Parallel()
	bus := &recBus{}
	store := newMemPushStore()
	sender := &stubSender{}
	if err := store.Upsert(t.Context(), model.PushSubscription{
		Endpoint: "https://push.example/a", P256dh: "p", Auth: "a",
	}); err != nil {
		t.Fatal(err)
	}
	f := &service.Fanout{SSE: bus, Push: service.NewPush(store, sender, "pub")}
	f.Notify(model.Notification{ID: "n1", Title: itemTitleDomain, ToStatus: model.StatusExpired})
	bus.mu.Lock()
	got := len(bus.got)
	bus.mu.Unlock()
	if got != 1 {
		t.Fatalf("sse %d", got)
	}
	sender.mu.Lock()
	n := sender.n
	sender.mu.Unlock()
	if n != 1 {
		t.Fatalf("push %d", n)
	}
}
