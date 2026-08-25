package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"duekeep/internal/db"
	"duekeep/internal/model"
)

// Kinds — SQL к item_kinds.
type Kinds struct {
	pool *pgxpool.Pool
}

// NewKinds создаёт репозиторий типов.
func NewKinds(pool *pgxpool.Pool) *Kinds {
	return &Kinds{pool: pool}
}

func (r *Kinds) q(ctx context.Context) db.Querier {
	return db.QuerierFrom(ctx, r.pool)
}

// List все типы, по slug.
func (r *Kinds) List(ctx context.Context) ([]model.Kind, error) {
	rows, err := r.q(ctx).Query(ctx, `
SELECT id::text, slug, name, color, attr_schema
FROM item_kinds ORDER BY slug`)
	if err != nil {
		return nil, fmt.Errorf("list kinds: %w", err)
	}
	defer rows.Close()
	out := make([]model.Kind, 0)
	for rows.Next() {
		k, err := scanKind(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// ByID ищет тип. Нет строки → ErrNotFound.
func (r *Kinds) ByID(ctx context.Context, id string) (model.Kind, error) {
	return r.scanOne(ctx, `
SELECT id::text, slug, name, color, attr_schema
FROM item_kinds WHERE id = $1::uuid`, id)
}

// Create вставляет kind. Конфликт slug → ErrConflict.
func (r *Kinds) Create(ctx context.Context, k model.Kind) (model.Kind, error) {
	schema, err := json.Marshal(k.AttrSchema)
	if err != nil {
		return model.Kind{}, err
	}
	created, err := r.scanOne(ctx, `
INSERT INTO item_kinds (slug, name, color, attr_schema)
VALUES ($1, $2, $3, $4)
RETURNING id::text, slug, name, color, attr_schema`, k.Slug, k.Name, k.Color, schema)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Kind{}, model.ErrConflict
		}
		return model.Kind{}, err
	}
	return created, nil
}

// Update пишет все поля уже собранной сущности.
func (r *Kinds) Update(ctx context.Context, k model.Kind) (model.Kind, error) {
	schema, err := json.Marshal(k.AttrSchema)
	if err != nil {
		return model.Kind{}, err
	}
	updated, err := r.scanOne(ctx, `
UPDATE item_kinds SET slug = $2, name = $3, color = $4, attr_schema = $5
WHERE id = $1::uuid
RETURNING id::text, slug, name, color, attr_schema`, k.ID, k.Slug, k.Name, k.Color, schema)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Kind{}, model.ErrConflict
		}
		return model.Kind{}, err
	}
	return updated, nil
}

// Delete удаляет строку. 0 rows → ErrNotFound.
func (r *Kinds) Delete(ctx context.Context, id string) error {
	tag, err := r.q(ctx).Exec(ctx, `DELETE FROM item_kinds WHERE id = $1::uuid`, id)
	if err != nil {
		return fmt.Errorf("delete kind: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

// CountItems заглушка: таблицы items ещё нет (Sprint 3).
func (r *Kinds) CountItems(context.Context, string) (int, error) {
	return 0, nil
}

type kindRow interface {
	Scan(dest ...any) error
}

func (r *Kinds) scanOne(ctx context.Context, q string, args ...any) (model.Kind, error) {
	k, err := scanKind(r.q(ctx).QueryRow(ctx, q, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Kind{}, model.ErrNotFound
	}
	if err != nil {
		return model.Kind{}, fmt.Errorf("kind row: %w", err)
	}
	return k, nil
}

func scanKind(row kindRow) (model.Kind, error) {
	var k model.Kind
	var raw []byte
	if err := row.Scan(&k.ID, &k.Slug, &k.Name, &k.Color, &raw); err != nil {
		return model.Kind{}, err
	}
	if len(raw) == 0 {
		k.AttrSchema = []model.AttrField{}
		return k, nil
	}
	if err := json.Unmarshal(raw, &k.AttrSchema); err != nil {
		return model.Kind{}, fmt.Errorf("attr_schema: %w", err)
	}
	if k.AttrSchema == nil {
		k.AttrSchema = []model.AttrField{}
	}
	return k, nil
}
