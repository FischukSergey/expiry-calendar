package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"duekeep/internal/clock"
)

// Run вставляет демо-данные. Повторный вызов не плодит строки; даты items обновляет.
func Run(ctx context.Context, pool *pgxpool.Pool, clk clock.Clock) error {
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
	if err := seedItems(ctx, pool, clk); err != nil {
		return fmt.Errorf("seed items: %w", err)
	}
	if err := seedRenewals(ctx, pool, clk); err != nil {
		return fmt.Errorf("seed renewals: %w", err)
	}
	if err := seedAudit(ctx, pool, clk); err != nil {
		return fmt.Errorf("seed audit: %w", err)
	}
	if err := seedNotifications(ctx, pool, clk); err != nil {
		return fmt.Errorf("seed notifications: %w", err)
	}
	slog.InfoContext(ctx, "seed completed")
	return nil
}

// EnsureKinds пишет справочник item_kinds инсталляции. Конфликт									 по slug — пропуск.
func EnsureKinds(ctx context.Context, pool *pgxpool.Pool) error {
	return seedKinds(ctx, pool)
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

// seedKinds пишет 10 типов. Конфликт по slug — пропуск (id/схема не меняются).
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
INSERT INTO categories (id, owner_id, parent_id, name, sort_order, created_at)
VALUES ($1, $2::uuid, NULLIF($3, '')::uuid, $4, $5, now())
ON CONFLICT (id) DO NOTHING`

	for _, c := range categorySeeds {
		if _, err := pool.Exec(ctx, q, c.id, adminID, c.parentID, c.name, c.sortOrder); err != nil {
			return fmt.Errorf("insert category %s: %w", c.name, err)
		}
	}
	return nil
}

// seedItems пишет каталог. Конфликт по id — обновляет даты и статус, не плодит строки.
func seedItems(ctx context.Context, pool *pgxpool.Pool, clk clock.Clock) error {
	const q = `
INSERT INTO items (
    id, owner_id, title, description, kind_id, category_id, vendor, tags,
    cost_amount, currency, billing_period, started_at, expires_at,
    notify_before_days, url, account_hint, status, attrs, created_at, updated_at
) VALUES (
    $1, $2::uuid, $3, $4, $5::uuid, NULLIF($6, '')::uuid, $7, $8,
    $9, $10, $11, $12, $13,
    $14, $15, $16, $17, $18, now(), now()
)
ON CONFLICT (id) DO UPDATE SET
    started_at = EXCLUDED.started_at,
    expires_at = EXCLUDED.expires_at,
    status = EXCLUDED.status,
    updated_at = now()`

	today := clock.Today(clk)
	for _, it := range itemSeeds() {
		kindID := kindIDBySlug(it.kindSlug)
		if kindID == "" {
			return fmt.Errorf("unknown kind slug %s", it.kindSlug)
		}
		started, expires := itemDates(today, it.startDays, it.expireDays)
		status := itemComputedStatus(today, it)
		attrs, err := marshalAttrs(it.attrs)
		if err != nil {
			return fmt.Errorf("attrs %s: %w", it.title, err)
		}
		if _, err := pool.Exec(ctx, q,
			it.id, adminID, it.title, it.description, kindID, it.categoryID, it.vendor, it.tags,
			it.cost, it.currency, it.billing, started, expires,
			it.notifyDays, it.url, it.account, status, attrs,
		); err != nil {
			return fmt.Errorf("insert item %s: %w", it.title, err)
		}
	}
	return nil
}

// seedRenewals пишет историю продлений. Конфликт по id — даты относительно today.
func seedRenewals(ctx context.Context, pool *pgxpool.Pool, clk clock.Clock) error {
	const q = `
INSERT INTO renewals (
    id, item_id, actor_id, old_expires_at, new_expires_at, old_cost, new_cost, comment, created_at
) VALUES (
    $1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, now()
)
ON CONFLICT (id) DO UPDATE SET
    old_expires_at = EXCLUDED.old_expires_at,
    new_expires_at = EXCLUDED.new_expires_at,
    old_cost = EXCLUDED.old_cost,
    new_cost = EXCLUDED.new_cost,
    comment = EXCLUDED.comment`

	today := clock.Today(clk)
	for _, r := range renewalSeeds() {
		it, ok := itemByN(r.itemN)
		if !ok {
			return fmt.Errorf("renewal %d: unknown item %d", r.n, r.itemN)
		}
		oldExp := today.AddDate(0, 0, r.oldExpire)
		newExp := today.AddDate(0, 0, r.newExpire)
		if _, err := pool.Exec(ctx, q,
			renewalID(r.n), it.id, adminID, oldExp, newExp, r.oldCost, r.newCost, r.comment,
		); err != nil {
			return fmt.Errorf("insert renewal %d: %w", r.n, err)
		}
	}
	return nil
}

// seedAudit пишет журнал. Конфликт по id — пропуск (снимок не освежаем).
func seedAudit(ctx context.Context, pool *pgxpool.Pool, clk clock.Clock) error {
	const q = `
INSERT INTO audit_log (id, owner_id, actor_id, action, entity, entity_id, before_json, after_json, created_at)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4, 'item', $5::uuid, $6, $7, now())
ON CONFLICT (id) DO NOTHING`

	today := clock.Today(clk)
	for _, a := range auditSeeds() {
		it, ok := itemByN(a.itemN)
		if !ok {
			return fmt.Errorf("audit %d: unknown item %d", a.n, a.itemN)
		}
		after, err := seedAuditAfter(today, it)
		if err != nil {
			return fmt.Errorf("audit snap %s: %w", it.title, err)
		}
		var before []byte
		if a.action == actionRenew {
			rs := renewalSeeds()
			if a.renewN < 1 || a.renewN > len(rs) {
				return fmt.Errorf("audit %d: bad renewal %d", a.n, a.renewN)
			}
			old := rs[a.renewN-1]
			before, err = json.Marshal(map[string]any{
				"id":          it.id,
				"title":       it.title,
				"kind_id":     kindIDBySlug(it.kindSlug),
				"expires_at":  dateOnly(today.AddDate(0, 0, old.oldExpire)),
				"cost_amount": old.oldCost,
			})
			if err != nil {
				return fmt.Errorf("audit before %s: %w", it.title, err)
			}
		}
		if _, err := pool.Exec(ctx, q, auditID(a.n), adminID, adminID, a.action, it.id, before, after); err != nil {
			return fmt.Errorf("insert audit %d: %w", a.n, err)
		}
	}
	return nil
}

// seedNotifications пишет unread для expired/expiring. Конфликт по id — снова unread.
func seedNotifications(ctx context.Context, pool *pgxpool.Pool, clk clock.Clock) error {
	const q = `
INSERT INTO notifications (id, owner_id, item_id, to_status, title, read_at, created_at)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, NULL, now())
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    to_status = EXCLUDED.to_status,
    item_id = EXCLUDED.item_id,
    created_at = EXCLUDED.created_at,
    read_at = NULL`

	today := clock.Today(clk)
	for _, it := range itemSeeds() {
		st := itemComputedStatus(today, it)
		if st != statusExpired && st != statusExpiring {
			continue
		}
		n := itemNFromID(it.id)
		if n < 1 {
			return fmt.Errorf("notification: bad item id %s", it.id)
		}
		if _, err := pool.Exec(ctx, q, noteID(n), adminID, it.id, st, it.title); err != nil {
			return fmt.Errorf("insert notification %s: %w", it.title, err)
		}
	}
	return nil
}

func itemNFromID(id string) int {
	var n int
	_, err := fmt.Sscanf(id, "55555555-5555-5555-5555-5555555555%d", &n)
	if err != nil {
		return 0
	}
	return n
}
