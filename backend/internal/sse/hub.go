package sse

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"

	"duekeep/internal/model"
)

// Имена SSE-событий по контракту Sprint 4.
const (
	EventNotification = "notification"
	EventPing         = "ping"
)

const clientBuf = 16

// Event — одна запись event-stream.
type Event struct {
	Name string
	Data []byte
}

type client struct {
	userID string
	ch     chan Event
}

// Hub — in-memory клиенты. Одна реплика процесса; горутинобезопасен.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]client
}

// NewHub пустой хаб.
func NewHub() *Hub {
	return &Hub{clients: map[string]client{}}
}

// Subscribe регистрирует клиента с sub. Канал не закрываем: отмена — через Unsubscribe + ctx.
func (h *Hub) Subscribe(userID string) (id string, ch <-chan Event) {
	id = uuid.NewString()
	c := make(chan Event, clientBuf)
	h.mu.Lock()
	h.clients[id] = client{userID: userID, ch: c}
	h.mu.Unlock()
	return id, c
}

// Unsubscribe снимает клиента. Повторный вызов безопасен.
func (h *Hub) Unsubscribe(id string) {
	h.mu.Lock()
	delete(h.clients, id)
	h.mu.Unlock()
}

// Publish шлёт всем. Полный буфер клиента пропускаем — тикер не должен стопориться.
func (h *Hub) Publish(ev Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		select {
		case c.ch <- ev:
		default:
		}
	}
}

// Notify — EventBus для тикера: только клиентам с тем же sub, что owner_id.
func (h *Hub) Notify(n model.Notification) {
	if n.OwnerID == "" {
		return
	}
	body, err := json.Marshal(struct {
		ID       string `json:"id"`
		ItemID   string `json:"item_id"`
		ToStatus string `json:"to_status"`
		Title    string `json:"title"`
	}{ID: n.ID, ItemID: n.ItemID, ToStatus: n.ToStatus, Title: n.Title})
	if err != nil {
		return
	}
	ev := Event{Name: EventNotification, Data: body}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		if c.userID != n.OwnerID {
			continue
		}
		select {
		case c.ch <- ev:
		default:
		}
	}
}
