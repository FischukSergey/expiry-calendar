package handler_test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"duekeep/internal/handler"
	"duekeep/internal/model"
	"duekeep/internal/sse"
)

func TestEventsRequiresAuth(t *testing.T) {
	t.Parallel()
	api, _, _ := notifyAPI(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/events", nil)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestEventsQueryTokenPing(t *testing.T) {
	t.Parallel()
	api, _, _ := notifyAPI(t)
	tok := testJWT(t, string(model.RoleViewer))
	name, data := openSSE(t, api, "", "/api/v1/events?access_token="+tok)
	if name != sse.EventPing || data != "{}" {
		t.Fatalf("first event %s %s", name, data)
	}
}

func TestEventsBearerPing(t *testing.T) {
	t.Parallel()
	api, _, _ := notifyAPI(t)
	tok := testJWT(t, string(model.RoleAdmin))
	name, data := openSSE(t, api, tok, "/api/v1/events")
	if name != sse.EventPing || data != "{}" {
		t.Fatalf("first event %s %s", name, data)
	}
}

func TestEventsPeriodicPing(t *testing.T) {
	t.Parallel()
	tok := testJWT(t, string(model.RoleViewer))
	pingAPI := handler.New(handler.Deps{
		Health: fakeHealth{}, Auth: fakeAuth{},
		Kinds: nopKinds{}, Categories: nopCategories{}, Items: nopItems{},
		Notifications: nopNotifications{}, Hub: sse.NewHub(),
		SSEPing: 20 * time.Millisecond, JWTSecret: []byte("handler-test-secret"),
	})
	srv := httptest.NewServer(pingAPI.Router())
	t.Cleanup(srv.Close)
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/v1/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	sc := bufio.NewScanner(resp.Body)
	if ev, _ := nextSSE(t, sc); ev != sse.EventPing {
		t.Fatalf("immediate %s", ev)
	}
	if ev, _ := nextSSE(t, sc); ev != sse.EventPing {
		t.Fatalf("periodic %s", ev)
	}
}

func TestEventsSeesTickerNotification(t *testing.T) {
	t.Parallel()
	api, tkr, store := notifyAPI(t)
	tok := testJWT(t, string(model.RoleViewer))
	_, err := store.Create(t.Context(), model.Item{
		OwnerID: fixtureUUID, Title: itemTitleDomain, KindID: otherKindID, Status: model.StatusActive,
		ExpiresAt: expiresSoon, NotifyBeforeDays: 30, Tags: []string{}, Attrs: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(api.Router())
	t.Cleanup(srv.Close)
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/api/v1/events?access_token="+tok, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		t.Fatalf("ct %s", resp.Header.Get("Content-Type"))
	}
	sc := bufio.NewScanner(resp.Body)
	if ev, _ := nextSSE(t, sc); ev != sse.EventPing {
		t.Fatalf("want ping got %s", ev)
	}
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	name, raw := nextSSE(t, sc)
	if name != sse.EventNotification {
		t.Fatalf("want notification got %s %s", name, raw)
	}
	var payload struct {
		ItemID   string `json:"item_id"`
		ToStatus string `json:"to_status"`
		Title    string `json:"title"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Title != itemTitleDomain || payload.ToStatus != model.StatusExpiring || payload.ItemID == "" {
		t.Fatalf("payload %+v", payload)
	}
}

func openSSE(t *testing.T, api *handler.API, bearer, path string) (name, data string) {
	t.Helper()
	srv := httptest.NewServer(api.Router())
	t.Cleanup(srv.Close)
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	return nextSSE(t, bufio.NewScanner(resp.Body))
}

func nextSSE(t *testing.T, sc *bufio.Scanner) (name, data string) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	type result struct {
		name string
		data string
	}
	ch := make(chan result, 1)
	go func() {
		var name, data string
		for sc.Scan() {
			line := sc.Text()
			switch {
			case strings.HasPrefix(line, "event: "):
				name = strings.TrimPrefix(line, "event: ")
			case strings.HasPrefix(line, "data: "):
				data = strings.TrimPrefix(line, "data: ")
			case line == "" && name != "":
				ch <- result{name, data}
				return
			}
		}
		if err := sc.Err(); err != nil && err != io.EOF {
			ch <- result{"", err.Error()}
		}
	}()
	select {
	case got := <-ch:
		if got.name == "" {
			t.Fatalf("sse read: %s", got.data)
		}
		return got.name, got.data
	case <-deadline:
		t.Fatal("sse timeout")
	}
	return "", ""
}
