package handler_test

import (
	"context"
	"sync"

	"duekeep/internal/model"
)

type memPush struct {
	mu   sync.Mutex
	byEP map[string]model.PushSubscription
}

func newMemPush() *memPush {
	return &memPush{byEP: map[string]model.PushSubscription{}}
}

func (m *memPush) Upsert(_ context.Context, s model.PushSubscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byEP[s.Endpoint] = s
	return nil
}

func (m *memPush) DeleteByEndpoint(_ context.Context, endpoint string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.byEP, endpoint)
	return nil
}

func (m *memPush) List(context.Context) ([]model.PushSubscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]model.PushSubscription, 0, len(m.byEP))
	for _, s := range m.byEP {
		out = append(out, s)
	}
	return out, nil
}

func (m *memPush) get(endpoint string) (model.PushSubscription, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.byEP[endpoint]
	return s, ok
}

type recSender struct {
	mu       sync.Mutex
	payloads [][]byte
	status   int
	err      error
}

func (s *recSender) Send(_ context.Context, _ model.PushSubscription, payload []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.payloads = append(s.payloads, append([]byte(nil), payload...))
	if s.err != nil {
		return 0, s.err
	}
	if s.status == 0 {
		return 201, nil
	}
	return s.status, nil
}

func (s *recSender) n() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.payloads)
}
