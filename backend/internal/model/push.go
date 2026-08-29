package model

import "time"

// PushSubscription — Web Push подписка браузера (один endpoint = одно устройство).
type PushSubscription struct {
	ID        string
	UserID    string
	Endpoint  string
	P256dh    string
	Auth      string
	UserAgent string
	CreatedAt time.Time
}

// PushSubscribe — тело POST /push/subscribe.
type PushSubscribe struct {
	Endpoint string   `json:"endpoint"`
	Keys     PushKeys `json:"keys"`
}

// PushKeys — ключи PushSubscription из браузера.
type PushKeys struct {
	P256dh string `json:"p256dh"`
	Auth   string `json:"auth"`
}

// PushUnsubscribe — тело DELETE /push/subscribe.
type PushUnsubscribe struct {
	Endpoint string `json:"endpoint"`
}

// VAPIDPublic — ответ GET /push/vapid-public.
type VAPIDPublic struct {
	PublicKey string `json:"public_key"`
}
