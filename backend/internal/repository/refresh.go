package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"duekeep/internal/db"
	"duekeep/internal/model"
)

// RefreshTokens — SQL к refresh_tokens. Сырой токен сюда не попадает.
type RefreshTokens struct {
	pool *pgxpool.Pool
}

// NewRefreshTokens создаёт репозиторий refresh.
func NewRefreshTokens(pool *pgxpool.Pool) *RefreshTokens {
	return &RefreshTokens{pool: pool}
}

func (r *RefreshTokens) q(ctx context.Context) db.Querier {
	return db.QuerierFrom(ctx, r.pool)
}

// Insert пишет новую сессию (хеш уникален).
func (r *RefreshTokens) Insert(ctx context.Context, rec model.RefreshSession) error {
	const q = `
INSERT INTO refresh_tokens (id, user_id, family_id, token_hash, expires_at, user_agent, created_at)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7)`
	_, err := r.q(ctx).Exec(ctx, q,
		rec.ID, rec.UserID, rec.FamilyID, rec.TokenHash, rec.ExpiresAt, rec.UserAgent, rec.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert refresh: %w", err)
	}
	return nil
}

// ByHash ищет сессию без блокировки (logout).
func (r *RefreshTokens) ByHash(ctx context.Context, hash string) (model.RefreshSession, error) {
	return r.scan(ctx, `
SELECT id::text, user_id::text, family_id::text, token_hash, expires_at, revoked_at, user_agent, created_at
FROM refresh_tokens WHERE token_hash = $1`, hash)
}

// ByHashForUpdate — тот же поиск с FOR UPDATE. Вызывать только внутри RunTx.
func (r *RefreshTokens) ByHashForUpdate(ctx context.Context, hash string) (model.RefreshSession, error) {
	return r.scan(ctx, `
SELECT id::text, user_id::text, family_id::text, token_hash, expires_at, revoked_at, user_agent, created_at
FROM refresh_tokens WHERE token_hash = $1 FOR UPDATE`, hash)
}

func (r *RefreshTokens) scan(ctx context.Context, q, hash string) (model.RefreshSession, error) {
	var rec model.RefreshSession
	err := r.q(ctx).QueryRow(ctx, q, hash).Scan(
		&rec.ID, &rec.UserID, &rec.FamilyID, &rec.TokenHash,
		&rec.ExpiresAt, &rec.RevokedAt, &rec.UserAgent, &rec.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.RefreshSession{}, model.ErrNotFound
	}
	if err != nil {
		return model.RefreshSession{}, fmt.Errorf("get refresh: %w", err)
	}
	return rec, nil
}

// RevokeID ставит revoked_at, если ещё не отозван.
func (r *RefreshTokens) RevokeID(ctx context.Context, id string, at time.Time) error {
	_, err := r.q(ctx).Exec(ctx, `
UPDATE refresh_tokens SET revoked_at = $2 WHERE id = $1::uuid AND revoked_at IS NULL`, id, at)
	if err != nil {
		return fmt.Errorf("revoke refresh: %w", err)
	}
	return nil
}

// RevokeFamily гасит все живые токены семьи (reuse).
func (r *RefreshTokens) RevokeFamily(ctx context.Context, familyID string, at time.Time) error {
	_, err := r.q(ctx).Exec(ctx, `
UPDATE refresh_tokens SET revoked_at = $2 WHERE family_id = $1::uuid AND revoked_at IS NULL`, familyID, at)
	if err != nil {
		return fmt.Errorf("revoke family: %w", err)
	}
	return nil
}

// RevokeUser гасит все сессии пользователя (logout-all).
func (r *RefreshTokens) RevokeUser(ctx context.Context, userID string, at time.Time) error {
	_, err := r.q(ctx).Exec(ctx, `
UPDATE refresh_tokens SET revoked_at = $2 WHERE user_id = $1::uuid AND revoked_at IS NULL`, userID, at)
	if err != nil {
		return fmt.Errorf("revoke user refresh: %w", err)
	}
	return nil
}
