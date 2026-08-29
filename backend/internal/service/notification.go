package service

import (
	"context"

	"duekeep/internal/model"
)

// Notification — лента и пометки прочитанного. Данные общие, без ролей.
type Notification struct {
	notes NotificationStore
}

// NewNotification собирает сервис ленты.
func NewNotification(notes NotificationStore) *Notification {
	return &Notification{notes: notes}
}

// List пагинирует. unread — только без read_at.
func (s *Notification) List(ctx context.Context, unread bool, page model.Page) (model.NotificationList, error) {
	rows, total, err := s.notes.List(ctx, unread, page)
	if err != nil {
		return model.NotificationList{}, err
	}
	if rows == nil {
		rows = []model.Notification{}
	}
	return model.NotificationList{Items: rows, Page: page.Page, PerPage: page.PerPage, Total: total}, nil
}

// MarkRead — POST /notifications/{id}/read. Повтор уже прочитанного — не ошибка.
func (s *Notification) MarkRead(ctx context.Context, id string) error {
	return s.notes.MarkRead(ctx, id)
}

// MarkAllRead — POST /notifications/read-all.
func (s *Notification) MarkAllRead(ctx context.Context) error {
	return s.notes.MarkAllRead(ctx)
}
