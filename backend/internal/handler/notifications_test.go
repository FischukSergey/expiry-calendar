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

func notifyAPI(t *testing.T) (*handler.API, *service.Ticker, *memItems) {
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
	hub := sse.NewHub()
	clk := clock.Fixed{T: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)}
	items := service.NewItem(store, kinds, newMemCats(), newMemRenewals(), newMemAudit(), nopTx, clk)
	api := handler.New(handler.Deps{
		Health:        fakeHealth{},
		Auth:          fakeAuth{},
		Kinds:         service.NewKind(kinds),
		Categories:    service.NewCategory(newMemCats()),
		Items:         items,
		Overview:      service.NewOverview(store, clk),
		Notifications: service.NewNotification(notes),
		Hub:           hub,
		JWTSecret:     []byte("handler-test-secret"),
		RefreshTTL:    336 * time.Hour,
	})
	return api, service.NewTicker(store, notes, nopTx, clk, hub), store
}

func TestNotificationsReadFlow(t *testing.T) {
	t.Parallel()
	api, tkr, store := notifyAPI(t)
	tok := testJWT(t, string(model.RoleViewer))
	created, err := store.Create(t.Context(), model.Item{
		Title: itemTitleDomain, KindID: otherKindID, Status: model.StatusActive,
		ExpiresAt: expiresSoon, NotifyBeforeDays: 30, Tags: []string{}, Attrs: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}

	list := listNotifications(t, api, tok, "")
	if list.Total != 1 || len(list.Items) != 1 {
		t.Fatalf("list %+v", list)
	}
	if list.Items[0].ItemID != created.ID || list.Items[0].ToStatus != model.StatusExpiring {
		t.Fatalf("item %+v", list.Items[0])
	}
	if list.Items[0].ReadAt != nil || list.Items[0].Title != itemTitleDomain {
		t.Fatalf("unread/title %+v", list.Items[0])
	}

	unread := listNotifications(t, api, tok, "unread=true")
	if unread.Total != 1 {
		t.Fatalf("unread %+v", unread)
	}

	adminJSON(t, api, tok, http.MethodPost, "/api/v1/notifications/"+list.Items[0].ID+"/read", "", http.StatusNoContent)
	if listNotifications(t, api, tok, "unread=true").Total != 0 {
		t.Fatal("still unread after read")
	}
	if listNotifications(t, api, tok, "").Total != 1 {
		t.Fatal("read hidden from full list")
	}

	second, err := store.Create(t.Context(), model.Item{
		Title: "Полис", KindID: otherKindID, Status: model.StatusExpiring,
		ExpiresAt: "2026-08-25", NotifyBeforeDays: 30, Tags: []string{}, Attrs: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	_ = second
	if listNotifications(t, api, tok, "unread=true").Total != 1 {
		t.Fatal("expected one unread after second tick")
	}
	adminJSON(t, api, tok, http.MethodPost, "/api/v1/notifications/read-all", "", http.StatusNoContent)
	if listNotifications(t, api, tok, "unread=true").Total != 0 {
		t.Fatal("read-all left unread")
	}
}

func TestTickerIntegrationStatusAndNotification(t *testing.T) {
	t.Parallel()
	api, tkr, store := notifyAPI(t)
	tok := testJWT(t, string(model.RoleViewer))
	created, err := store.Create(t.Context(), model.Item{
		Title: itemTitleDomain, KindID: otherKindID, Status: model.StatusActive,
		ExpiresAt: expiresSoon, NotifyBeforeDays: 30, Tags: []string{}, Attrs: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tkr.Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	got, err := store.ByID(t.Context(), created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.StatusExpiring {
		t.Fatalf("status %s", got.Status)
	}
	list := listNotifications(t, api, tok, "")
	if list.Total != 1 || list.Items[0].ToStatus != model.StatusExpiring || list.Items[0].ItemID != created.ID {
		t.Fatalf("notes %+v", list)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard %d %s", rec.Code, rec.Body.String())
	}
	var dash model.Dashboard
	if err := json.NewDecoder(rec.Body).Decode(&dash); err != nil {
		t.Fatal(err)
	}
	if len(dash.ExpirationsByMonth) != 6 || len(dash.Soonest) != 1 {
		t.Fatalf("series %+v", dash)
	}
	if dash.Soonest[0].ID != created.ID || dash.Soonest[0].Status != model.StatusExpiring {
		t.Fatalf("soonest %+v", dash.Soonest)
	}
}

func TestNotificationsRequiresAuth(t *testing.T) {
	t.Parallel()
	api, _, _ := notifyAPI(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/notifications", nil)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func listNotifications(t *testing.T, api *handler.API, tok, query string) model.NotificationList {
	t.Helper()
	path := "/api/v1/notifications"
	if query != "" {
		path += "?" + query
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list %d %s", rec.Code, rec.Body.String())
	}
	var out model.NotificationList
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}
