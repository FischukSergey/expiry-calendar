package model

import (
	"encoding/json"
	"time"
)

// DateLayout — календарная дата в API (гггг-мм-дд).
const DateLayout = "2006-01-02"

// Статусы записи и связанные константы API.
const (
	StatusActive    = "active"
	StatusExpiring  = "expiring"
	StatusExpired   = "expired"
	StatusCancelled = "cancelled"
	StatusArchived  = "archived"

	BillingOneTime = "one_time"
	BillingMonthly = "monthly"
	BillingYearly  = "yearly"

	CurrencyRUB = "RUB"

	DefaultNotifyDays = 30
	DefaultPerPage    = 20
	MaxPerPage        = 100

	SortExpiresAt  = "expires_at"
	SortCostAmount = "cost_amount"
	SortTitle      = "title"
	SortUpdatedAt  = "updated_at"

	OrderAsc  = "asc"
	OrderDesc = "desc"

	AuditEntityItem = "item"
	AuditCreate     = "create"
	AuditUpdate     = "update"
	AuditDelete     = "delete"
	AuditRenew      = "renew"
	AuditBulk       = "bulk"
)

// Item — запись истечения. Даты started_at/expires_at — DateLayout.
type Item struct {
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	KindID           string         `json:"kind_id"`
	CategoryID       *string        `json:"category_id"`
	Vendor           string         `json:"vendor"`
	Tags             []string       `json:"tags"`
	CostAmount       int            `json:"cost_amount"`
	Currency         string         `json:"currency"`
	BillingPeriod    string         `json:"billing_period"`
	StartedAt        *string        `json:"started_at"`
	ExpiresAt        string         `json:"expires_at"`
	NotifyBeforeDays int            `json:"notify_before_days"`
	URL              string         `json:"url"`
	AccountHint      string         `json:"account_hint"`
	Status           string         `json:"status"`
	Attrs            map[string]any `json:"attrs"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// ItemList — GET /items.
type ItemList struct {
	Items   []Item `json:"items"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Total   int    `json:"total"`
}

// ItemCard — GET /items/{id}.
type ItemCard struct {
	Item     Item      `json:"item"`
	Renewals []Renewal `json:"renewals"`
}

// ItemFilter — query GET /items. CategoryIDs заполняет service (узел + потомки).
type ItemFilter struct {
	Q             string
	KindID        string
	Status        string
	CategoryID    string
	CategoryIDs   []string
	Vendor        string
	ExpiresFrom   string
	ExpiresTo     string
	CostFrom      *int
	CostTo        *int
	BillingPeriod string
	Tag           string
	Sort          string
	Order         string
}

// Page — offset-пагинация после нормализации.
type Page struct {
	Page    int
	PerPage int
}

// Offset для LIMIT/OFFSET.
func (p Page) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// ItemPatch — частичное обновление. Set* отличает «не прислали» от JSON null.
type ItemPatch struct {
	Title            *string
	Description      *string
	KindID           *string
	SetCategory      bool
	CategoryID       *string
	Vendor           *string
	Tags             *[]string
	CostAmount       *int
	Currency         *string
	BillingPeriod    *string
	SetStarted       bool
	StartedAt        *string
	ExpiresAt        *string
	NotifyBeforeDays *int
	URL              *string
	AccountHint      *string
	Status           *string
	SetAttrs         bool
	Attrs            map[string]any
}

// RenewInput — тело POST /items/{id}/renew.
type RenewInput struct {
	NewExpiresAt string
	NewCost      *int
	Comment      string
}

// BulkInput — тело POST /items/bulk.
type BulkInput struct {
	IDs        []string
	CategoryID *string
	Status     *string
}

// BulkResult — ответ bulk.
type BulkResult struct {
	Updated int `json:"updated"`
}

// Renewal — строка истории продления.
type Renewal struct {
	ID           string    `json:"id"`
	ItemID       string    `json:"item_id"`
	ActorID      string    `json:"actor_id"`
	OldExpiresAt string    `json:"old_expires_at"`
	NewExpiresAt string    `json:"new_expires_at"`
	OldCost      int       `json:"old_cost"`
	NewCost      int       `json:"new_cost"`
	Comment      string    `json:"comment"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuditEntry — строка audit_log.
type AuditEntry struct {
	ID         string          `json:"id"`
	ActorID    *string         `json:"actor_id"`
	Action     string          `json:"action"`
	Entity     string          `json:"entity"`
	EntityID   string          `json:"entity_id"`
	BeforeJSON json.RawMessage `json:"before_json"`
	AfterJSON  json.RawMessage `json:"after_json"`
	CreatedAt  time.Time       `json:"created_at"`
}

// AuditList — GET /audit.
type AuditList struct {
	Items   []AuditEntry `json:"items"`
	Page    int          `json:"page"`
	PerPage int          `json:"per_page"`
	Total   int          `json:"total"`
}
