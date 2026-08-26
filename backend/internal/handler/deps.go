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

// KindService — справочник типов.
type KindService interface {
	List(ctx context.Context) ([]model.Kind, error)
	Create(ctx context.Context, in model.Kind) (model.Kind, error)
	Patch(ctx context.Context, id string, p model.KindPatch) (model.Kind, error)
	Delete(ctx context.Context, id string) error
}

// CategoryService — дерево категорий.
type CategoryService interface {
	List(ctx context.Context) ([]model.Category, error)
	Create(ctx context.Context, parentID *string, name string, sortOrder int) (model.Category, error)
	Patch(ctx context.Context, id string, p model.CategoryPatch) (model.Category, error)
	Delete(ctx context.Context, id string) error
}

// ItemService — записи, renew, bulk, audit list.
type ItemService interface {
	List(ctx context.Context, f model.ItemFilter, page model.Page) (model.ItemList, error)
	Create(ctx context.Context, in model.Item, actorID string) (model.Item, error)
	Get(ctx context.Context, id string) (model.ItemCard, error)
	Patch(ctx context.Context, id string, p model.ItemPatch, actorID string) (model.Item, error)
	Delete(ctx context.Context, id, actorID string) error
	Renew(ctx context.Context, id string, in model.RenewInput, actorID string) (model.Item, error)
	Bulk(ctx context.Context, in model.BulkInput, actorID string) (model.BulkResult, error)
	ListAudit(ctx context.Context, page model.Page) (model.AuditList, error)
}

// Deps — зависимости HTTP-слоя. JWTSecret нужен middleware, не service повторно.
type Deps struct {
	Health       HealthService
	Auth         AuthService
	Kinds        KindService
	Categories   CategoryService
	Items        ItemService
	Spec         []byte
	JWTSecret    []byte
	CookieSecure bool
	RefreshTTL   time.Duration
}
