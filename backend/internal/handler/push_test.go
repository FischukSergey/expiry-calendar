package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"duekeep/internal/clock"
	"duekeep/internal/handler"
	"duekeep/internal/model"
	"duekeep/internal/service"
	"duekeep/internal/sse"
)

const (
	testVAPIDPublic = "test-vapid-public"
	pushEndpoint    = "https://push.example/sub/1"
	pushBody        = `{"endpoint":"https://push.example/sub/1","keys":{"p256dh":"p256","auth":"authk"}}`
)

func pushAPI(t *testing.T) (*handler.API, *memPush, *recSender, *service.Ticker, *memItems) {
	t.Helper()
	kinds := newMemKinds()
	_, err := kinds.Create(t.Context(), model.Kind{
		Slug: kindSlugOther, Name: kindNameOther, Color: kindColorBlack, AttrSchema: []model.AttrField{},
	})
	if err != nil {
		t.Fatal(err)
	}
	kinds.mu.Lock()
	k := kinds.byID[fixtureUUID]
	k.ID = otherKindID
	delete(kinds.byID, fixtureUUID)
	kinds.byID[otherKindID] = k
	kinds.mu.Unlock()

	store := newMemItems()
	notes := newMemNotifications()
	subs := newMemPush()
	sender := &recSender{}
	push := service.NewPush(subs, sender, testVAPIDPublic)
	hub := sse.NewHub()
	clk := clock.Fixed{T: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)}
	items := service.NewItem(store, kinds, newMemCats(), newMemRenewals(), newMemAudit(), nopTx, clk)
	api := handler.New(handler.Deps{
		Health:        fakeHealth{},
		Auth:          fakeAuth{},
		Kinds:         service.NewKind(kinds),
		Categories:    service.NewCategory(newMemCats()),
		Items:         items,
		Notifications: service.NewNotification(notes),
		Push:          push,
		Hub:           hub,
		JWTSecret:     []byte("handler-test-secret"),
		RefreshTTL:    336 * time.Hour,
	})
	tkr := service.NewTicker(store, notes, nopTx, clk, &service.Fanout{SSE: hub, Push: push})
	return api, subs, sender, tkr, store
}

func TestPushVAPIDPublic(t *testing.T) {
	t.Parallel()
	api, _, _, _, _ := pushAPI(t)
	tok := testJWT(t, string(model.RoleViewer))
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/push/vapid-public", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var out model.VAPIDPublic
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.PublicKey != testVAPIDPublic {
		t.Fatalf("key %q", out.PublicKey)
	}
}

func TestPushRequiresAuth(t *testing.T) {
	t.Parallel()
	api, _, _, _, _ := pushAPI(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/push/vapid-public", nil)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestPushSubscribeUpsertAndDelete(t *testing.T) {
	t.Parallel()
	api, subs, _, _, _ := pushAPI(t)
	tok := testJWT(t, string(model.RoleViewer))
	adminJSON(t, api, tok, http.MethodPost, "/api/v1/push/subscribe", pushBody, http.StatusNoContent)
	got, ok := subs.get(pushEndpoint)
	if !ok || got.P256dh != "p256" || got.Auth != "authk" {
		t.Fatalf("stored %+v ok=%v", got, ok)
	}
	if got.UserID != fixtureUUID {
		t.Fatalf("user %s", got.UserID)
	}
	adminJSON(t, api, tok, http.MethodPost, "/api/v1/push/subscribe",
		`{"endpoint":"https://push.example/sub/1","keys":{"p256dh":"p256-2","auth":"auth-2"}}`,
		http.StatusNoContent)
	got, ok = subs.get(pushEndpoint)
	if !ok || got.P256dh != "p256-2" || got.Auth != "auth-2" {
		t.Fatalf("upsert %+v", got)
	}
	subs.mu.Lock()
	n := len(subs.byEP)
	subs.mu.Unlock()
	if n != 1 {
		t.Fatalf("want 1 sub got %d", n)
	}
	adminJSON(t, api, tok, http.MethodDelete, "/api/v1/push/subscribe",
		`{"endpoint":"https://push.example/sub/1"}`, http.StatusNoContent)
	if _, ok = subs.get(pushEndpoint); ok {
		t.Fatal("still stored")
	}
	if _, ok = subs.get("https://push.example/missing"); ok {
		t.Fatal("missing endpoint")
	}
	adminJSON(t, api, tok, http.MethodDelete, "/api/v1/push/subscribe",
		`{"endpoint":"https://push.example/sub/1"}`, http.StatusNoContent)
}

func TestPushSubscribeValidation(t *testing.T) {
	t.Parallel()
	api, _, _, _, _ := pushAPI(t)
	tok := testJWT(t, string(model.RoleViewer))
	adminJSON(t, api, tok, http.MethodPost, "/api/v1/push/subscribe",
		`{"endpoint":"","keys":{"p256dh":"x","auth":"y"}}`, http.StatusUnprocessableEntity)
	adminJSON(t, api, tok, http.MethodDelete, "/api/v1/push/subscribe",
		`{"endpoint":""}`, http.StatusUnprocessableEntity)
}

func TestPushTickerBroadcastAnd410(t *testing.T) {
	t.Parallel()
	api, subs, sender, tkr, store := pushAPI(t)
	tok := testJWT(t, string(model.RoleViewer))
	adminJSON(t, api, tok, http.MethodPost, "/api/v1/push/subscribe", pushBody, http.StatusNoContent)
	_, err := store.Create(t.Context(), model.Item{
		OwnerID: fixtureUUID, Title: itemTitleDomain, KindID: otherKindID, Status: model.StatusActive,
		ExpiresAt: expiresSoon, NotifyBeforeDays: 30, Tags: []string{}, Attrs: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if sender.n() != 1 {
		t.Fatalf("sends %d", sender.n())
	}
	var payload struct {
		Title    string `json:"title"`
		ToStatus string `json:"to_status"`
	}
	sender.mu.Lock()
	raw := sender.payloads[0]
	sender.mu.Unlock()
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Title != itemTitleDomain || payload.ToStatus != model.StatusExpiring {
		t.Fatalf("payload %+v", payload)
	}

	sender.mu.Lock()
	sender.status = http.StatusGone
	sender.mu.Unlock()
	_, err = store.Create(t.Context(), model.Item{
		OwnerID: fixtureUUID, Title: "Полис", KindID: otherKindID, Status: model.StatusExpiring,
		ExpiresAt: "2026-08-25", NotifyBeforeDays: 30, Tags: []string{}, Attrs: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, ok := subs.get(pushEndpoint); ok {
		t.Fatal("410 did not drop subscription")
	}
}
