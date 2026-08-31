package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"duekeep/internal/db"
	"duekeep/internal/model"
)

// Categories — SQL к categories. Дерево собирает service.
type Categories struct {
	pool *pgxpool.Pool
}

// NewCategories создаёт репозиторий категорий.
func NewCategories(pool *pgxpool.Pool) *Categories {
	return &Categories{pool: pool}
}

func (r *Categories) q(ctx context.Context) db.Querier {
	return db.QuerierFrom(ctx, r.pool)
}

// List плоский список: sort_order, name. Children пустые.
func (r *Categories) List(ctx context.Context) ([]model.Category, error) {
	rows, err := r.q(ctx).Query(ctx, `
SELECT id::text, owner_id::text, parent_id::text, name, sort_order
FROM categories
ORDER BY sort_order, name`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()
	out := make([]model.Category, 0)
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ByID одна строка без детей.
func (r *Categories) ByID(ctx context.Context, id string) (model.Category, error) {
	c, err := scanCategory(r.q(ctx).QueryRow(ctx, `
SELECT id::text, owner_id::text, parent_id::text, name, sort_order
FROM categories WHERE id = $1::uuid`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Category{}, model.ErrNotFound
	}
	if err != nil {
		return model.Category{}, fmt.Errorf("get category: %w", err)
	}
	return c, nil
}

// Create вставляет узел. parent_id NULL — корень.
func (r *Categories) Create(ctx context.Context, c model.Category) (model.Category, error) {
	created, err := scanCategory(r.q(ctx).QueryRow(ctx, `
INSERT INTO categories (owner_id, parent_id, name, sort_order)
VALUES ($1::uuid, $2, $3, $4)
RETURNING id::text, owner_id::text, parent_id::text, name, sort_order`, c.OwnerID, parentArg(c.ParentID), c.Name, c.SortOrder))
	if err != nil {
		return model.Category{}, fmt.Errorf("insert category: %w", err)
	}
	return created, nil
}

// Update пишет parent/name/sort. 0 rows → ErrNotFound.
func (r *Categories) Update(ctx context.Context, c model.Category) (model.Category, error) {
	updated, err := scanCategory(r.q(ctx).QueryRow(ctx, `
UPDATE categories SET parent_id = $2, name = $3, sort_order = $4
WHERE id = $1::uuid
RETURNING id::text, owner_id::text, parent_id::text, name, sort_order`, c.ID, parentArg(c.ParentID), c.Name, c.SortOrder))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Category{}, model.ErrNotFound
	}
	if err != nil {
		return model.Category{}, fmt.Errorf("update category: %w", err)
	}
	return updated, nil
}

// Delete удаляет узел. 0 rows → ErrNotFound.
func (r *Categories) Delete(ctx context.Context, id string) error {
	tag, err := r.q(ctx).Exec(ctx, `DELETE FROM categories WHERE id = $1::uuid`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

// CountChildren — сколько прямых детей (запрет DELETE).
func (r *Categories) CountChildren(ctx context.Context, id string) (int, error) {
	var n int
	err := r.q(ctx).QueryRow(ctx, `SELECT count(*) FROM categories WHERE parent_id = $1::uuid`, id).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count children: %w", err)
	}
	return n, nil
}

// DescendantIDs — id узла и всех потомков (CTE). Нет узла → пустой список.
func (r *Categories) DescendantIDs(ctx context.Context, id string) ([]string, error) {
	rows, err := r.q(ctx).Query(ctx, `
WITH RECURSIVE tree AS (
    SELECT id FROM categories WHERE id = $1::uuid
    UNION ALL
    SELECT c.id FROM categories c JOIN tree t ON c.parent_id = t.id
)
SELECT id::text FROM tree`, id)
	if err != nil {
		return nil, fmt.Errorf("category descendants: %w", err)
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var one string
		if err := rows.Scan(&one); err != nil {
			return nil, err
		}
		out = append(out, one)
	}
	return out, rows.Err()
}

// CountItems — сколько записей с этим category_id (запрет DELETE занятой категории).
func (r *Categories) CountItems(ctx context.Context, id string) (int, error) {
	var n int
	err := r.q(ctx).QueryRow(ctx, `SELECT count(*) FROM items WHERE category_id = $1::uuid`, id).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count items by category: %w", err)
	}
	return n, nil
}

func parentArg(id *string) any {
	if id == nil || *id == "" {
		return nil
	}
	return *id
}

type catRow interface {
	Scan(dest ...any) error
}

func scanCategory(row catRow) (model.Category, error) {
	var c model.Category
	var parent *string
	if err := row.Scan(&c.ID, &c.OwnerID, &parent, &c.Name, &c.SortOrder); err != nil {
		return model.Category{}, err
	}
	c.ParentID = parent
	c.Children = []model.Category{}
	return c, nil
}
