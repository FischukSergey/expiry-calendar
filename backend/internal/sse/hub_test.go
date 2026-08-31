package sse_test

import (
	"sync"
	"testing"

	"duekeep/internal/model"
	"duekeep/internal/sse"
)

func TestHubPublishAndUnsubscribe(t *testing.T) {
	t.Parallel()
	h := sse.NewHub()
	id, ch := h.Subscribe("u1")
	h.Publish(sse.Event{Name: sse.EventPing, Data: []byte("{}")})
	got := <-ch
	if got.Name != sse.EventPing {
		t.Fatalf("event %s", got.Name)
	}
	h.Unsubscribe(id)
	h.Publish(sse.Event{Name: sse.EventPing, Data: []byte("{}")})
	select {
	case <-ch:
		t.Fatal("got event after unsubscribe")
	default:
	}
}

func TestHubConcurrent(t *testing.T) {
	t.Parallel()
	h := sse.NewHub()
	const n = 32
	var wg sync.WaitGroup
	wg.Add(n * 2)
	ids := make([]string, n)
	for i := range n {
		id, ch := h.Subscribe("u")
		ids[i] = id
		go func(_ <-chan sse.Event) {
			defer wg.Done()
			for range 8 {
				h.Publish(sse.Event{Name: sse.EventPing, Data: []byte("{}")})
			}
		}(ch)
		go func(ch <-chan sse.Event) {
			defer wg.Done()
			for range 4 {
				_, ok := <-ch
				if !ok {
					return
				}
			}
		}(ch)
	}
	wg.Wait()
	for _, id := range ids {
		h.Unsubscribe(id)
	}
	h.Notify(model.Notification{OwnerID: "u", ID: "1", ItemID: "2", ToStatus: model.StatusExpiring, Title: "x"})
}

func TestHubNotifyOnlyOwner(t *testing.T) {
	t.Parallel()
	h := sse.NewHub()
	_, mine := h.Subscribe("owner")
	_, theirs := h.Subscribe("other")
	h.Notify(model.Notification{
		OwnerID: "owner", ID: "1", ItemID: "2", ToStatus: model.StatusExpiring, Title: "x",
	})
	select {
	case ev := <-mine:
		if ev.Name != sse.EventNotification {
			t.Fatalf("name %s", ev.Name)
		}
	default:
		t.Fatal("owner missed event")
	}
	select {
	case <-theirs:
		t.Fatal("leaked to other")
	default:
	}
}

func TestHubNotifyJSON(t *testing.T) {
	t.Parallel()
	h := sse.NewHub()
	_, ch := h.Subscribe("owner")
	h.Notify(model.Notification{
		OwnerID:  "owner",
		ID:       "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		ItemID:   "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		ToStatus: model.StatusExpired, Title: "Полис",
	})
	got := <-ch
	if got.Name != sse.EventNotification {
		t.Fatalf("name %s", got.Name)
	}
	const want = `{"id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",` +
		`"item_id":"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb","to_status":"expired","title":"Полис"}`
	if string(got.Data) != want {
		t.Fatalf("data %s", got.Data)
	}
}
