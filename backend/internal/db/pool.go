package db

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Connect открывает пул pgx и проверяет ping.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgx ping: %w", err)
	}
	return pool, nil
}

// Migrate прогоняет goose.Up из embed FS.
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrationsFS fs.FS, dir string) error {
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer func() {
		if err := sqlDB.Close(); err != nil {
			slog.WarnContext(ctx, "close sql db after migrate", "err", err)
		}
	}()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("sql ping: %w", err)
	}

	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, sqlDB, dir); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

// Close закрывает пул, если он создан.
func Close(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}
