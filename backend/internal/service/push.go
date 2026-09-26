package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
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

// hostLookup резолвит хост push-endpoint. В тестах подменяется на публичный адрес.
type hostLookup func(ctx context.Context, host string) ([]net.IP, error)

// Push — subscribe/unsubscribe и рассылка из тикера.
type Push struct {
	store  PushStore
	sender PushSender
	public string
	lookup hostLookup
}

// NewPush собирает сервис. sender nil — Broadcast ничего не шлёт (тесты без HTTP).
func NewPush(store PushStore, sender PushSender, publicKey string) *Push {
	return &Push{store: store, sender: sender, public: publicKey, lookup: defaultLookup}
}

// AllowPublicHosts считает любой хост публичным. Только для тестов без DNS.
func (s *Push) AllowPublicHosts() {
	s.lookup = func(context.Context, string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("1.1.1.1")}, nil
	}
}

func defaultLookup(ctx context.Context, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, "ip", host)
}

// PublicKey — GET /push/vapid-public.
func (s *Push) PublicKey() string {
	return s.public
}

// Subscribe — POST /push/subscribe, upsert по endpoint.
func (s *Push) Subscribe(ctx context.Context, userID string, in model.PushSubscribe, userAgent string) error {
	sub, err := validateSubscribe(ctx, s.lookup, in)
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

// Broadcast шлёт подпискам владельца item. 410/404/403 — удаляем строку. Ошибка одного endpoint не стопорит остальных.
func (s *Push) Broadcast(ctx context.Context, n model.Notification) error {
	if s.sender == nil || n.OwnerID == "" {
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
		if sub.UserID != n.OwnerID {
			continue
		}
		if err := endpointPublic(ctx, s.lookup, sub.Endpoint); err != nil {
			slog.ErrorContext(ctx, "push skip private endpoint", "err", err)
			continue
		}
		status, sendErr := s.sender.Send(ctx, sub, payload)
		if sendErr != nil {
			slog.ErrorContext(ctx, "push send", "err", sendErr, "status", status)
		} else if status < 200 || status >= 300 {
			slog.ErrorContext(ctx, "push send", "status", status)
		}
		if subscriptionDead(status) {
			if delErr := s.store.DeleteByEndpoint(ctx, sub.Endpoint); delErr != nil {
				slog.ErrorContext(ctx, "push drop dead", "err", delErr, "status", status)
			}
		}
	}
	return nil
}

// subscriptionDead — endpoint больше не принимает VAPID (410/404/403).
func subscriptionDead(status int) bool {
	return status == http.StatusGone || status == http.StatusNotFound || status == http.StatusForbidden
}

func validateSubscribe(ctx context.Context, lookup hostLookup, in model.PushSubscribe) (model.PushSubscription, error) {
	endpoint := strings.TrimSpace(in.Endpoint)
	p256dh := strings.TrimSpace(in.Keys.P256dh)
	auth := strings.TrimSpace(in.Keys.Auth)
	fields := map[string]any{}
	if endpoint == "" {
		fields[fieldEndpoint] = detailRequired
	} else if err := endpointPublic(ctx, lookup, endpoint); err != nil {
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

// endpointPublic пускает только https на глобальный unicast-адрес. Частные и loopback — нет.
func endpointPublic(ctx context.Context, lookup hostLookup, raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" {
		return errPushEndpoint
	}
	host := u.Hostname()
	if strings.EqualFold(host, "localhost") {
		return errPushEndpoint
	}
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicIP(ip) {
			return errPushEndpoint
		}
		return nil
	}
	if lookup == nil {
		lookup = defaultLookup
	}
	ips, err := lookup(ctx, host)
	if err != nil || len(ips) == 0 {
		return errPushEndpoint
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return errPushEndpoint
		}
	}
	return nil
}

func isPublicIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	addr = addr.Unmap()
	return addr.IsGlobalUnicast() && !addr.IsPrivate()
}

var errPushEndpoint = model.Validation("invalid endpoint", map[string]any{fieldEndpoint: "url"})
