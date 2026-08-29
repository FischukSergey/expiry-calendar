package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"duekeep/internal/model"
)

const (
	fieldEndpoint = "endpoint"
	fieldP256dh   = "keys.p256dh"
	fieldAuthKey  = "keys.auth"
	pushTTL       = 86400
)

// PushStore — таблица подписок.
type PushStore interface {
	Upsert(ctx context.Context, s model.PushSubscription) error
	DeleteByEndpoint(ctx context.Context, endpoint string) error
	List(ctx context.Context) ([]model.PushSubscription, error)
}

// PushSender — один POST к push-сервису. 410 обрабатывает вызывающий, не sender.
type PushSender interface {
	Send(ctx context.Context, sub model.PushSubscription, payload []byte) (status int, err error)
}

// Push — subscribe/unsubscribe и рассылка из тикера.
type Push struct {
	store  PushStore
	sender PushSender
	public string
}

// NewPush собирает сервис. sender nil — Broadcast ничего не шлёт (тесты без HTTP).
func NewPush(store PushStore, sender PushSender, publicKey string) *Push {
	return &Push{store: store, sender: sender, public: publicKey}
}

// PublicKey — GET /push/vapid-public.
func (s *Push) PublicKey() string {
	return s.public
}

// Subscribe — POST /push/subscribe, upsert по endpoint.
func (s *Push) Subscribe(ctx context.Context, userID string, in model.PushSubscribe, userAgent string) error {
	sub, err := validateSubscribe(in)
	if err != nil {
		return err
	}
	sub.UserID = userID
	sub.UserAgent = strings.Clone(userAgent)
	return s.store.Upsert(ctx, sub)
}

// Unsubscribe — DELETE /push/subscribe. Нет строки — не ошибка.
func (s *Push) Unsubscribe(ctx context.Context, endpoint string) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return model.Validation("invalid endpoint", map[string]any{fieldEndpoint: detailRequired})
	}
	return s.store.DeleteByEndpoint(ctx, endpoint)
}

// Broadcast шлёт всем подпискам. 410 — удаляем строку. Ошибка одного endpoint не стопорит остальных.
func (s *Push) Broadcast(ctx context.Context, n model.Notification) error {
	if s.sender == nil {
		return nil
	}
	subs, err := s.store.List(ctx)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(struct {
		ID       string `json:"id"`
		ItemID   string `json:"item_id"`
		ToStatus string `json:"to_status"`
		Title    string `json:"title"`
	}{ID: n.ID, ItemID: n.ItemID, ToStatus: n.ToStatus, Title: n.Title})
	if err != nil {
		return err
	}
	for _, sub := range subs {
		status, sendErr := s.sender.Send(ctx, sub, payload)
		if sendErr != nil {
			slog.ErrorContext(ctx, "push send", "err", sendErr)
			continue
		}
		if status == http.StatusGone {
			if delErr := s.store.DeleteByEndpoint(ctx, sub.Endpoint); delErr != nil {
				slog.ErrorContext(ctx, "push drop 410", "err", delErr)
			}
		}
	}
	return nil
}

func validateSubscribe(in model.PushSubscribe) (model.PushSubscription, error) {
	endpoint := strings.TrimSpace(in.Endpoint)
	p256dh := strings.TrimSpace(in.Keys.P256dh)
	auth := strings.TrimSpace(in.Keys.Auth)
	fields := map[string]any{}
	if endpoint == "" {
		fields[fieldEndpoint] = detailRequired
	} else if u, err := url.Parse(endpoint); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		fields[fieldEndpoint] = "url"
	}
	if p256dh == "" {
		fields[fieldP256dh] = detailRequired
	}
	if auth == "" {
		fields[fieldAuthKey] = detailRequired
	}
	if len(fields) > 0 {
		return model.PushSubscription{}, model.Validation("invalid subscription", fields)
	}
	return model.PushSubscription{
		Endpoint: strings.Clone(endpoint),
		P256dh:   strings.Clone(p256dh),
		Auth:     strings.Clone(auth),
	}, nil
}
