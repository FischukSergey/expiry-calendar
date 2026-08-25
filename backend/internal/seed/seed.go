package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Run вставляет демо-пользователей и справочники. Повторный вызов не плодит строки.
func Run(ctx context.Context, pool *pgxpool.Pool) error {
	if err := CheckCatalog(); err != nil {
		return fmt.Errorf("seed catalog: %w", err)
	}
	if err := seedUsers(ctx, pool); err != nil {
		return fmt.Errorf("seed users: %w", err)
	}
	if err := seedKinds(ctx, pool); err != nil {
		return fmt.Errorf("seed kinds: %w", err)
	}
	if err := seedCategories(ctx, pool); err != nil {
		return fmt.Errorf("seed categories: %w", err)
	}
	slog.InfoContext(ctx, "seed completed")
	return nil
}

// seedUsers пишет admin и viewer. Конфликт по email — пропуск, пароль не обновляет.
func seedUsers(ctx context.Context, pool *pgxpool.Pool) error {
	const q = `
INSERT INTO users (id, email, password_hash, role, created_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (email) DO NOTHING`

	for _, u := range userSeeds {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("bcrypt %s: %w", u.email, err)
		}
		if _, err := pool.Exec(ctx, q, u.id, u.email, string(hash), u.role); err != nil {
			return fmt.Errorf("insert user %s: %w", u.email, err)
		}
	}
	return nil
}

// seedKinds пишет 9 типов. Конфликт по slug — пропуск (id/схема не меняются).
func seedKinds(ctx context.Context, pool *pgxpool.Pool) error {
	const q = `
INSERT INTO item_kinds (id, slug, name, color, attr_schema, created_at)
VALUES ($1, $2, $3, $4, $5, now())
ON CONFLICT (slug) DO NOTHING`

	for _, k := range kindSeeds {
		schema, err := json.Marshal(k.attrSchema)
		if err != nil {
			return fmt.Errorf("attr_schema %s: %w", k.slug, err)
		}
		if _, err := pool.Exec(ctx, q, k.id, k.slug, k.name, k.color, schema); err != nil {
			return fmt.Errorf("insert kind %s: %w", k.slug, err)
		}
	}
	return nil
}

// seedCategories пишет дерево. Пустой parent_id в SQL → NULL; конфликт по id.
func seedCategories(ctx context.Context, pool *pgxpool.Pool) error {
	const q = `
INSERT INTO categories (id, parent_id, name, sort_order, created_at)
VALUES ($1, NULLIF($2, '')::uuid, $3, $4, now())
ON CONFLICT (id) DO NOTHING`

	for _, c := range categorySeeds {
		if _, err := pool.Exec(ctx, q, c.id, c.parentID, c.name, c.sortOrder); err != nil {
			return fmt.Errorf("insert category %s: %w", c.name, err)
		}
	}
	return nil
}
