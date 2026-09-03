package handler

import (
	"context"
	"time"

	"duekeep/internal/model"
	"duekeep/internal/sse"
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
	List(ctx context.Context, ownerID string) ([]model.Category, error)
	Create(ctx context.Context, parentID *string, name string, sortOrder int, ownerID string) (model.Category, error)
	Patch(ctx context.Context, id string, p model.CategoryPatch, actorID string) (model.Category, error)
	Delete(ctx context.Context, id, actorID string) error
}

// ItemService — записи, renew, bulk, audit list.
type ItemService interface {
	List(ctx context.Context, f model.ItemFilter, page model.Page, actorID string) (model.ItemList, error)
	Create(ctx context.Context, in model.Item, actorID string) (model.Item, error)
	Get(ctx context.Context, id, actorID string) (model.ItemCard, error)
	Patch(ctx context.Context, id string, p model.ItemPatch, actorID string) (model.Item, error)
	Delete(ctx context.Context, id, actorID string) error
	Renew(ctx context.Context, id string, in model.RenewInput, actorID string) (model.Item, error)
	Bulk(ctx context.Context, in model.BulkInput, actorID string) (model.BulkResult, error)
	ListAudit(ctx context.Context, page model.Page, actorID string) (model.AuditList, error)
	Pay(ctx context.Context, id, date, actorID string) (model.ItemPayment, bool, error)
	Unpay(ctx context.Context, id, date, actorID string) error
	Export(ctx context.Context, f model.ItemFilter, actorID string) ([]byte, error)
	Import(
		ctx context.Context,
		csvData []byte,
		mapping map[string]string,
		dryRun bool,
		actorID string,
	) (model.CSVImportPreview, model.CSVImportResult, error)
}

// OverviewService — дашборд и календарь.
type OverviewService interface {
	Dashboard(ctx context.Context, ownerID string) (model.Dashboard, error)
	Calendar(ctx context.Context, year, month int, ownerID string) (model.Calendar, error)
}

// NotificationService — лента и read / read-all.
type NotificationService interface {
	List(ctx context.Context, ownerID string, unread bool, page model.Page) (model.NotificationList, error)
	MarkRead(ctx context.Context, id, ownerID string) error
	MarkAllRead(ctx context.Context, ownerID string) error
}

// PushService — VAPID и подписки Web Push.
type PushService interface {
	PublicKey() string
	Subscribe(ctx context.Context, userID string, in model.PushSubscribe, userAgent string) error
	Unsubscribe(ctx context.Context, endpoint string) error
}

// Deps — зависимости HTTP-слоя. JWTSecret нужен middleware, не service повторно.
type Deps struct {
	Health        HealthService
	Auth          AuthService
	Kinds         KindService
	Categories    CategoryService
	Items         ItemService
	Overview      OverviewService
	Notifications NotificationService
	Push          PushService
	Hub           *sse.Hub
	SSEPing       time.Duration // 0 → 15 с; в тестах можно укоротить.
	Spec          []byte
	JWTSecret     []byte
	CookieSecure  bool
	RefreshTTL    time.Duration
}
