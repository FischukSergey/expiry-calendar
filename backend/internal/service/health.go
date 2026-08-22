package service

import "context"

// HealthStore — доступ к проверке БД.
type HealthStore interface {
	Ping(ctx context.Context) error
}

// Health — сценарий liveness/readiness.
type Health struct {
	store HealthStore
}

// NewHealth собирает сервис health.
func NewHealth(store HealthStore) *Health {
	return &Health{store: store}
}

// Check пингует хранилище.
func (s *Health) Check(ctx context.Context) error {
	return s.store.Ping(ctx)
}
