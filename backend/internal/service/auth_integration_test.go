//go:build integration

package service_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"duekeep/internal/clock"
	"duekeep/internal/db"
	"duekeep/internal/model"
	"duekeep/internal/repository"
	"duekeep/internal/service"
	"duekeep/migrations"
)

func TestRefreshReuseCommitsOnPostgres(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Fatal("DATABASE_URL is required")
	}
	ctx := t.Context()
	pool, err := db.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close(pool) })
	if err := db.Migrate(ctx, pool, migrations.FS, "."); err != nil {
		t.Fatal(err)
	}

	users := repository.NewUsers(pool)
	tokens := repository.NewRefreshTokens(pool)
	runTx := func(ctx context.Context, fn func(context.Context) error) error {
		return db.RunTx(ctx, pool, fn)
	}
	auth := service.NewAuth(users, tokens, runTx, clock.Real{}, service.AuthConfig{
		Secret:     []byte("integration-test-secret"),
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 336 * time.Hour,
		BcryptCost: 4,
	})
	email := fmt.Sprintf("reuse-%d@duekeep.local", time.Now().UnixNano())
	first, err := auth.Register(ctx, email, "secret12", "test")
	if err != nil {
		t.Fatal(err)
	}
	second, err := auth.Refresh(ctx, first.RefreshToken, "test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := auth.Refresh(ctx, first.RefreshToken, "test"); !errors.Is(err, model.ErrUnauthorized) {
		t.Fatalf("replay old: %v", err)
	}
	if _, err := auth.Refresh(ctx, second.RefreshToken, "test"); !errors.Is(err, model.ErrUnauthorized) {
		t.Fatalf("newer token still accepted after reuse: %v", err)
	}
}
