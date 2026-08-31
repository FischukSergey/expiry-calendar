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
)

func overviewAPI(t *testing.T) (*handler.API, *memItems) {
	t.Helper()
	store := newMemItems()
	clk := clock.Fixed{T: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)}
	api := handler.New(handler.Deps{
		Health:        fakeHealth{},
		Auth:          fakeAuth{},
		Kinds:         nopKinds{},
		Categories:    nopCategories{},
		Items:         nopItems{},
		Overview:      service.NewOverview(store, clk),
		Notifications: nopNotifications{},
		JWTSecret:     []byte("handler-test-secret"),
		RefreshTTL:    336 * time.Hour,
	})
	return api, store
}

func TestDashboardRequiresAuth(t *testing.T) {
	t.Parallel()
	api, _ := overviewAPI(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/dashboard", nil)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestDashboardViewerTwoCurrencies(t *testing.T) {
	t.Parallel()
	api, store := overviewAPI(t)
	if _, err := store.Create(t.Context(), model.Item{
		OwnerID: fixtureUUID, Title: itemTitleDomain, KindID: otherKindID, Status: model.StatusActive,
		ExpiresAt: expiresSoon, CostAmount: 200, Currency: model.CurrencyRUB,
		BillingPeriod: model.BillingMonthly, Tags: []string{}, Attrs: map[string]any{},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(t.Context(), model.Item{
		OwnerID: fixtureUUID, Title: "SaaS", KindID: otherKindID, Status: model.StatusActive,
		ExpiresAt: "2026-12-01", CostAmount: 36, Currency: currencyUSD,
		BillingPeriod: model.BillingYearly, Tags: []string{}, Attrs: map[string]any{},
	}); err != nil {
		t.Fatal(err)
	}
	tok := testJWT(t, string(model.RoleViewer))
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var out model.Dashboard
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Counts.Active != 2 || len(out.UpcomingCost) != 2 {
		t.Fatalf("dash %+v", out)
	}
	rub, usd := out.UpcomingCost[0], out.UpcomingCost[1]
	if rub.Currency != model.CurrencyRUB || usd.Currency != currencyUSD {
		t.Fatalf("mixed currencies %+v", out.UpcomingCost)
	}
	if rub.Monthly != 200 || rub.Yearly != 2400 || usd.Yearly != 36 || usd.Monthly != 3 {
		t.Fatalf("converted or merged %+v", out.UpcomingCost)
	}
}

func TestCalendarQueryAndEmptyMonth(t *testing.T) {
	t.Parallel()
	api, store := overviewAPI(t)
	if _, err := store.Create(t.Context(), model.Item{
		OwnerID: fixtureUUID, Title: itemTitleDomain, KindID: otherKindID, Status: model.StatusExpiring,
		ExpiresAt: "2026-08-21", Tags: []string{}, Attrs: map[string]any{},
	}); err != nil {
		t.Fatal(err)
	}
	tok := testJWT(t, string(model.RoleViewer))
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/calendar?year=2026&month=8", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var out model.Calendar
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Year != 2026 || out.Month != 8 || len(out.Days) != 1 || out.Days[0].Date != "2026-08-21" {
		t.Fatalf("cal %+v", out)
	}
	empty := adminJSON(t, api, tok, http.MethodGet, "/api/v1/calendar?year=2026&month=7", "", http.StatusOK)
	var none model.Calendar
	if err := json.NewDecoder(empty.Body).Decode(&none); err != nil {
		t.Fatal(err)
	}
	if len(none.Days) != 0 {
		t.Fatalf("empty month %+v", none)
	}
	adminJSON(t, api, tok, http.MethodGet, "/api/v1/calendar?year=2026&month=13", "", http.StatusUnprocessableEntity)
	adminJSON(t, api, tok, http.MethodGet, "/api/v1/calendar", "", http.StatusUnprocessableEntity)
}

func TestDashboardOwnerIsolation(t *testing.T) {
	t.Parallel()
	api, store := overviewAPI(t)
	if _, err := store.Create(t.Context(), model.Item{
		OwnerID: fixtureUUID, Title: itemTitleDomain, KindID: otherKindID, Status: model.StatusActive,
		ExpiresAt: expiresSoon, CostAmount: 200, Currency: model.CurrencyRUB,
		BillingPeriod: model.BillingMonthly, Tags: []string{}, Attrs: map[string]any{},
	}); err != nil {
		t.Fatal(err)
	}
	other := testJWTSub(t, string(model.RoleAdmin), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+other)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var out model.Dashboard
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Counts.Active != 0 || len(out.Soonest) != 0 {
		t.Fatalf("leaked %+v", out)
	}
}
