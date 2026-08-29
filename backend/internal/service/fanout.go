package service

import (
	"context"
	"log/slog"
	"time"

	"duekeep/internal/model"
)

const pushBroadcastTimeout = 15 * time.Second

// Fanout — EventBus: SSE сразу, затем Web Push. Ошибка push не роняет тикер.
type Fanout struct {
	SSE  EventBus
	Push *Push
}

// Notify сначала SSE (буфер, без сети), потом push всем подпискам.
func (f *Fanout) Notify(n model.Notification) {
	if f.SSE != nil {
		f.SSE.Notify(n)
	}
	if f.Push == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), pushBroadcastTimeout)
	defer cancel()
	if err := f.Push.Broadcast(ctx, n); err != nil {
		slog.Error("push broadcast", "err", err)
	}
}
