package service

import (
	"encoding/json"

	"duekeep/internal/model"
)

// itemAuditSnap — краткий before/after. Нет url, account_hint, паролей.
type itemAuditSnap struct {
	ID         string         `json:"id"`
	Title      string         `json:"title"`
	KindID     string         `json:"kind_id"`
	CategoryID *string        `json:"category_id"`
	Status     string         `json:"status"`
	ExpiresAt  string         `json:"expires_at"`
	CostAmount int            `json:"cost_amount"`
	Attrs      map[string]any `json:"attrs"`
}

func itemSnap(it model.Item) json.RawMessage {
	b, err := json.Marshal(itemAuditSnap{
		ID: it.ID, Title: it.Title, KindID: it.KindID, CategoryID: it.CategoryID,
		Status: it.Status, ExpiresAt: it.ExpiresAt, CostAmount: it.CostAmount, Attrs: it.Attrs,
	})
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}

func auditEntry(actorID, action, entityID string, before, after json.RawMessage) model.AuditEntry {
	e := model.AuditEntry{
		Action:     action,
		Entity:     model.AuditEntityItem,
		EntityID:   entityID,
		BeforeJSON: before,
		AfterJSON:  after,
	}
	if actorID != "" {
		e.ActorID = &actorID
	}
	return e
}
