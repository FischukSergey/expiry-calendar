package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"duekeep/internal/db"
	"duekeep/internal/model"
)

// Users — SQL к таблице users.
type Users struct {
	pool *pgxpool.Pool
}

// NewUsers создаёт репозиторий пользователей.
func NewUsers(pool *pgxpool.Pool) *Users {
	return &Users{pool: pool}
}

func (r *Users) q(ctx context.Context) db.Querier {
	return db.QuerierFrom(ctx, r.pool)
}

// Create вставляет пользователя. Конфликт email → ErrConflict.
func (r *Users) Create(ctx context.Context, email, passwordHash string, role model.Role) (model.User, error) {
	const q = `
INSERT INTO users (email, password_hash, role)
VALUES ($1, $2, $3)
RETURNING id::text, email, password_hash, role, created_at`

	var u model.User
	err := r.q(ctx).QueryRow(ctx, q, email, passwordHash, string(role)).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, model.ErrConflict
		}
		return model.User{}, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

// ByEmail ищет по citext-email. Нет строки → ErrNotFound.
func (r *Users) ByEmail(ctx context.Context, email string) (model.User, error) {
	return r.scanUser(ctx, `
SELECT id::text, email, password_hash, role, created_at
FROM users WHERE email = $1`, email)
}

// ByID ищет по UUID.
func (r *Users) ByID(ctx context.Context, id string) (model.User, error) {
	return r.scanUser(ctx, `
SELECT id::text, email, password_hash, role, created_at
FROM users WHERE id = $1::uuid`, id)
}

func (r *Users) scanUser(ctx context.Context, q string, arg any) (model.User, error) {
	var u model.User
	err := r.q(ctx).QueryRow(ctx, q, arg).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, model.ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}
