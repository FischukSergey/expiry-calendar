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

// Hub — in-memory клиенты. Одна реплика процесса; горутинобезопасен.
type Hub struct {
	mu      sync.RWMutex
	clients map[string]chan Event
}

// NewHub пустой хаб.
func NewHub() *Hub {
	return &Hub{clients: map[string]chan Event{}}
}

// Subscribe регистрирует клиента. Канал не закрываем: отмена — через Unsubscribe + ctx.
func (h *Hub) Subscribe() (id string, ch <-chan Event) {
	id = uuid.NewString()
	c := make(chan Event, clientBuf)
	h.mu.Lock()
	h.clients[id] = c
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
	for _, ch := range h.clients {
		select {
		case ch <- ev:
		default:
		}
	}
}

// Notify — EventBus для тикера: только свежий INSERT, без read_at.
func (h *Hub) Notify(n model.Notification) {
	body, err := json.Marshal(struct {
		ID       string `json:"id"`
		ItemID   string `json:"item_id"`
		ToStatus string `json:"to_status"`
		Title    string `json:"title"`
	}{ID: n.ID, ItemID: n.ItemID, ToStatus: n.ToStatus, Title: n.Title})
	if err != nil {
		return
	}
	h.Publish(Event{Name: EventNotification, Data: body})
}
