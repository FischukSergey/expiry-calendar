package handler

import (
	"context"
	"time"

	"duekeep/internal/model"
)

// AuthService — сценарии токенов. Cookie собирает handler.
type AuthService interface {
	Register(ctx context.Context, email, password, userAgent string) (model.TokenPair, error)
	Login(ctx context.Context, email, password, userAgent string) (model.TokenPair, error)
	Refresh(ctx context.Context, raw, userAgent string) (model.TokenPair, error)
	Logout(ctx context.Context, raw string) error
	LogoutAll(ctx context.Context, userID string) error
	Me(ctx context.Context, userID string) (model.PublicUser, error)
}

// Deps — зависимости HTTP-слоя. JWTSecret нужен middleware, не service повторно.
type Deps struct {
	Health       HealthService
	Auth         AuthService
	Spec         []byte
	JWTSecret    []byte
	CookieSecure bool
	RefreshTTL   time.Duration
}
