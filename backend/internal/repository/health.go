package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Health ходит в БД проверкой SELECT 1.
type Health struct {
	pool *pgxpool.Pool
}

// NewHealth создаёт репозиторий health.
func NewHealth(pool *pgxpool.Pool) *Health {
	return &Health{pool: pool}
}

// Ping выполняет SELECT 1.
func (r *Health) Ping(ctx context.Context) error {
	var one int
	return r.pool.QueryRow(ctx, "SELECT 1").Scan(&one)
}
