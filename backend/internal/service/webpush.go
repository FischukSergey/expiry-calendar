package service

import (
	"context"
	"io"
	"strings"

	webpush "github.com/SherClockHolmes/webpush-go"

	"duekeep/internal/model"
)

// WebPushSender — реальная рассылка через webpush-go.
type WebPushSender struct {
	public  string
	private string
	subject string
	client  webpush.HTTPClient
}

// NewWebPushSender собирает sender. subject без mailto: — библиотека добавит сама.
func NewWebPushSender(public, private, subject string) *WebPushSender {
	return &WebPushSender{
		public:  public,
		private: private,
		subject: normalizeVAPIDSubject(subject),
	}
}

// Send один POST. Вызывающий смотрит status (410 → удалить подписку).
func (s *WebPushSender) Send(ctx context.Context, sub model.PushSubscription, payload []byte) (int, error) {
	opts := &webpush.Options{
		Subscriber:      s.subject,
		TTL:             pushTTL,
		VAPIDPublicKey:  s.public,
		VAPIDPrivateKey: s.private,
	}
	if s.client != nil {
		opts.HTTPClient = s.client
	}
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys:     webpush.Keys{Auth: sub.Auth, P256dh: sub.P256dh},
	}, opts)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
		_, _ = io.Copy(io.Discard, resp.Body)
		return resp.StatusCode, err
	}
	return 0, err
}

func normalizeVAPIDSubject(raw string) string {
	s := strings.TrimSpace(raw)
	if rest, ok := strings.CutPrefix(s, "mailto:"); ok {
		return rest
	}
	return s
}
