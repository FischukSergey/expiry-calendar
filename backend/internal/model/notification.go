package model

import "time"

// Notification — in-app событие перехода в expiring/expired.
type Notification struct {
	ID        string     `json:"id"`
	ItemID    string     `json:"item_id"`
	ToStatus  string     `json:"to_status"`
	Title     string     `json:"title"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// NotificationList — GET /notifications.
type NotificationList struct {
	Items   []Notification `json:"items"`
	Page    int            `json:"page"`
	PerPage int            `json:"per_page"`
	Total   int            `json:"total"`
}
