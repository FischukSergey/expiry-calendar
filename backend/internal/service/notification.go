package service

import (
	"context"

	"duekeep/internal/model"
)

// Notification — лента и пометки прочитанного. Только свой owner_id.
type Notification struct {
	notes NotificationStore
}

// NewNotification собирает сервис ленты.
func NewNotification(notes NotificationStore) *Notification {
	return &Notification{notes: notes}
}

// List пагинирует. unread — только без read_at.
func (s *Notification) List(ctx context.Context, ownerID string, unread bool, page model.Page) (model.NotificationList, error) {
	rows, total, err := s.notes.List(ctx, ownerID, unread, page)
	if err != nil {
		return model.NotificationList{}, err
	}
	if rows == nil {
		rows = []model.Notification{}
	}
	return model.NotificationList{Items: rows, Page: page.Page, PerPage: page.PerPage, Total: total}, nil
}

// MarkRead — POST /notifications/{id}/read. Чужой id → 404. Повтор — не ошибка.
func (s *Notification) MarkRead(ctx context.Context, id, ownerID string) error {
	return s.notes.MarkRead(ctx, id, ownerID)
}

// MarkAllRead — POST /notifications/read-all.
func (s *Notification) MarkAllRead(ctx context.Context, ownerID string) error {
	return s.notes.MarkAllRead(ctx, ownerID)
}
