package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"duekeep/internal/db"
	"duekeep/internal/model"
)

// Items — SQL к items.
type Items struct {
	pool *pgxpool.Pool
}

// NewItems создаёт репозиторий записей.
func NewItems(pool *pgxpool.Pool) *Items {
	return &Items{pool: pool}
}

func (r *Items) q(ctx context.Context) db.Querier {
	return db.QuerierFrom(ctx, r.pool)
}

const itemCols = `id::text, owner_id::text, title, description, kind_id::text, category_id::text, vendor, tags,
    cost_amount, trim(currency), billing_period, started_at, expires_at,
    notify_before_days, url, account_hint, status, attrs, created_at, updated_at`

// Create вставляет запись.
func (r *Items) Create(ctx context.Context, it model.Item) (model.Item, error) {
	attrs, err := marshalItemAttrs(it.Attrs)
	if err != nil {
		return model.Item{}, err
	}
	return r.scanOne(ctx, `
INSERT INTO items (
    owner_id, title, description, kind_id, category_id, vendor, tags,
    cost_amount, currency, billing_period, started_at, expires_at,
    notify_before_days, url, account_hint, status, attrs
) VALUES (
    $1::uuid, $2, $3, $4::uuid, NULLIF($5, '')::uuid, $6, $7,
    $8, $9, $10, $11, $12,
    $13, $14, $15, $16, $17
)
RETURNING `+itemCols, it.OwnerID, it.Title, it.Description, it.KindID, categoryArg(it.CategoryID), it.Vendor, it.Tags,
		it.CostAmount, it.Currency, it.BillingPeriod, dateArg(it.StartedAt), it.ExpiresAt,
		it.NotifyBeforeDays, it.URL, it.AccountHint, it.Status, attrs)
}

// ByID одна строка. Нет → ErrNotFound.
func (r *Items) ByID(ctx context.Context, id string) (model.Item, error) {
	return r.scanOne(ctx, `SELECT `+itemCols+` FROM items WHERE id = $1::uuid`, id)
}

// Update пишет все поля уже собранной сущности.
func (r *Items) Update(ctx context.Context, it model.Item) (model.Item, error) {
	attrs, err := marshalItemAttrs(it.Attrs)
	if err != nil {
		return model.Item{}, err
	}
	return r.scanOne(ctx, `
UPDATE items SET
    title = $2, description = $3, kind_id = $4::uuid, category_id = NULLIF($5, '')::uuid,
    vendor = $6, tags = $7, cost_amount = $8, currency = $9, billing_period = $10,
    started_at = $11, expires_at = $12, notify_before_days = $13, url = $14,
    account_hint = $15, status = $16, attrs = $17, updated_at = now()
WHERE id = $1::uuid
RETURNING `+itemCols, it.ID, it.Title, it.Description, it.KindID, categoryArg(it.CategoryID),
		it.Vendor, it.Tags, it.CostAmount, it.Currency, it.BillingPeriod,
		dateArg(it.StartedAt), it.ExpiresAt, it.NotifyBeforeDays, it.URL,
		it.AccountHint, it.Status, attrs)
}

// Delete удаляет строку. 0 rows → ErrNotFound.
func (r *Items) Delete(ctx context.Context, id string) error {
	tag, err := r.q(ctx).Exec(ctx, `DELETE FROM items WHERE id = $1::uuid`, id)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

// List фильтрует по колонкам и owner_id. CategoryIDs уже с потомками.
func (r *Items) List(ctx context.Context, f model.ItemFilter, page model.Page) ([]model.Item, int, error) {
	if f.OwnerID == "" {
		return []model.Item{}, 0, nil
	}
	where, args := itemWhere(f)
	var total int
	countQ := `SELECT count(*) FROM items` + where
	if err := r.q(ctx).QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count items: %w", err)
	}
	sortCol := map[string]string{
		model.SortExpiresAt:  "expires_at",
		model.SortCostAmount: "cost_amount",
		model.SortTitle:      "title",
		model.SortUpdatedAt:  "updated_at",
	}[f.Sort]
	if sortCol == "" {
		sortCol = "expires_at"
	}
	dir := "ASC"
	if f.Order == model.OrderDesc {
		dir = "DESC"
	}
	lim := len(args) + 1
	off := len(args) + 2
	args = append(args, page.PerPage, page.Offset())
	q := fmt.Sprintf(`SELECT %s FROM items%s ORDER BY %s %s, id LIMIT $%d OFFSET $%d`,
		itemCols, where, sortCol, dir, lim, off)
	rows, err := r.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()
	out := make([]model.Item, 0)
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

// BulkUpdate пишет category_id и/или status. Неизвестные и чужие id пропускает.
func (r *Items) BulkUpdate(ctx context.Context, ids []string, categoryID *string, status *string, ownerID string) (int, error) {
	if ownerID == "" {
		return 0, nil
	}
	tag, err := r.q(ctx).Exec(ctx, `
UPDATE items SET
    category_id = COALESCE(NULLIF($2, '')::uuid, category_id),
    status = COALESCE(NULLIF($3, ''), status),
    updated_at = now()
WHERE id = ANY($1::uuid[]) AND owner_id = $4::uuid`, ids, categoryArg(categoryID), statusArg(status), ownerID)
	if err != nil {
		return 0, fmt.Errorf("bulk items: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

func itemWhere(f model.ItemFilter) (string, []any) {
	var parts []string
	var args []any
	add := func(clause string, val any) {
		args = append(args, val)
		parts = append(parts, fmt.Sprintf(clause, len(args)))
	}
	add("owner_id = $%d::uuid", f.OwnerID)
	if f.Q != "" {
		pat := "%" + f.Q + "%"
		args = append(args, pat)
		n := len(args)
		parts = append(parts, fmt.Sprintf(
			"(title ILIKE $%d OR vendor ILIKE $%d OR EXISTS (SELECT 1 FROM unnest(tags) t WHERE t ILIKE $%d))",
			n, n, n))
	}
	if f.KindID != "" {
		add("kind_id = $%d::uuid", f.KindID)
	}
	if f.Status != "" {
		add("status = $%d", f.Status)
	}
	if len(f.CategoryIDs) > 0 {
		add("category_id = ANY($%d::uuid[])", f.CategoryIDs)
	}
	if f.Vendor != "" {
		add("vendor ILIKE $%d", f.Vendor)
	}
	if f.ExpiresFrom != "" {
		add("expires_at >= $%d", f.ExpiresFrom)
	}
	if f.ExpiresTo != "" {
		add("expires_at <= $%d", f.ExpiresTo)
	}
	if f.CostFrom != nil {
		add("cost_amount >= $%d", *f.CostFrom)
	}
	if f.CostTo != nil {
		add("cost_amount <= $%d", *f.CostTo)
	}
	if f.BillingPeriod != "" {
		add("billing_period = $%d", f.BillingPeriod)
	}
	if f.Tag != "" {
		add("$%d = ANY(tags)", f.Tag)
	}
	if len(parts) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

func (r *Items) scanOne(ctx context.Context, q string, args ...any) (model.Item, error) {
	it, err := scanItem(r.q(ctx).QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Item{}, model.ErrNotFound
	}
	if err != nil {
		return model.Item{}, fmt.Errorf("item row: %w", err)
	}
	return it, nil
}

type itemRow interface {
	Scan(dest ...any) error
}

func scanItem(row itemRow) (model.Item, error) {
	var it model.Item
	var cat *string
	var started *time.Time
	var expires time.Time
	var raw []byte
	var tags []string
	err := row.Scan(
		&it.ID, &it.OwnerID, &it.Title, &it.Description, &it.KindID, &cat, &it.Vendor, &tags,
		&it.CostAmount, &it.Currency, &it.BillingPeriod, &started, &expires,
		&it.NotifyBeforeDays, &it.URL, &it.AccountHint, &it.Status, &raw,
		&it.CreatedAt, &it.UpdatedAt,
	)
	if err != nil {
		return model.Item{}, err
	}
	it.CategoryID = cat
	it.Tags = tags
	if it.Tags == nil {
		it.Tags = []string{}
	}
	if started != nil {
		s := started.UTC().Format(model.DateLayout)
		it.StartedAt = &s
	}
	it.ExpiresAt = expires.UTC().Format(model.DateLayout)
	it.Attrs, err = unmarshalItemAttrs(raw)
	if err != nil {
		return model.Item{}, err
	}
	return it, nil
}

func marshalItemAttrs(attrs map[string]any) ([]byte, error) {
	if attrs == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(attrs)
}

func unmarshalItemAttrs(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("attrs: %w", err)
	}
	if m == nil {
		return map[string]any{}, nil
	}
	return m, nil
}

func categoryArg(id *string) string {
	if id == nil {
		return ""
	}
	return *id
}

func dateArg(s *string) any {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}

func statusArg(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ListOpen — все открытые записи для тикера (не cancelled/archived).
func (r *Items) ListOpen(ctx context.Context) ([]model.Item, error) {
	return r.listOpen(ctx, `SELECT `+itemCols+`
FROM items
WHERE status NOT IN ('cancelled', 'archived')
ORDER BY id`)
}

// ListOpenByOwner — открытые записи владельца для dashboard/calendar.
func (r *Items) ListOpenByOwner(ctx context.Context, ownerID string) ([]model.Item, error) {
	if ownerID == "" {
		return []model.Item{}, nil
	}
	return r.listOpen(ctx, `SELECT `+itemCols+`
FROM items
WHERE status NOT IN ('cancelled', 'archived') AND owner_id = $1::uuid
ORDER BY id`, ownerID)
}

func (r *Items) listOpen(ctx context.Context, q string, args ...any) ([]model.Item, error) {
	rows, err := r.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list open items: %w", err)
	}
	defer rows.Close()
	out := make([]model.Item, 0)
	for rows.Next() {
		it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// SetStatus пишет только status. cancelled/archived не обновляет (0 rows → ErrNotFound).
func (r *Items) SetStatus(ctx context.Context, id, status string) (model.Item, error) {
	it, err := r.scanOne(ctx, `
UPDATE items SET status = $2, updated_at = now()
WHERE id = $1::uuid AND status NOT IN ('cancelled', 'archived')
RETURNING `+itemCols, id, status)
	if err != nil {
		return model.Item{}, err
	}
	return it, nil
}
