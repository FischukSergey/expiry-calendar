package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	duekeep "duekeep"
	"duekeep/internal/handler"
)

func docsAPI() *handler.API {
	return handler.New(handler.Deps{
		Health:        fakeHealth{},
		Auth:          fakeAuth{},
		Kinds:         nopKinds{},
		Categories:    nopCategories{},
		Items:         nopItems{},
		Notifications: nopNotifications{},
		Spec:          duekeep.OpenAPISpec,
		JWTSecret:     []byte("x"),
	})
}

func TestOpenAPISpec(t *testing.T) {
	t.Parallel()
	api := docsAPI()
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/openapi.yaml", nil)
	api.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "/healthz") || !strings.Contains(body, "/api/v1/auth/login") {
		t.Fatalf("spec without health/auth: %s", body)
	}
	if !strings.Contains(body, "/api/v1/kinds") || !strings.Contains(body, "/api/v1/categories") {
		t.Fatalf("spec without catalogs: %s", body)
	}
	if !strings.Contains(body, "/api/v1/items") || !strings.Contains(body, "/api/v1/audit") {
		t.Fatalf("spec without items/audit: %s", body)
	}
	if !strings.Contains(body, "/api/v1/items/export") || !strings.Contains(body, "/api/v1/items/import") {
		t.Fatalf("spec without csv: %s", body)
	}
	if !strings.Contains(body, "/api/v1/notifications") {
		t.Fatalf("spec without notifications: %s", body)
	}
	if !strings.Contains(body, "/api/v1/events") {
		t.Fatalf("spec without events: %s", body)
	}
	if !strings.Contains(body, "/api/v1/push/vapid-public") || !strings.Contains(body, "/api/v1/push/subscribe") {
		t.Fatalf("spec without push: %s", body)
	}
	if !strings.Contains(body, "/api/v1/dashboard") || !strings.Contains(body, "/api/v1/calendar") {
		t.Fatalf("spec without overview: %s", body)
	}
	if !strings.Contains(body, "duekeep_refresh") || !strings.Contains(body, "bearerAuth") {
		t.Fatalf("spec without refresh cookie / bearerAuth")
	}
	for _, path := range []string{
		"/api/v1/auth/register", "/api/v1/auth/login", "/api/v1/auth/refresh",
		"/api/v1/auth/logout", "/api/v1/auth/logout-all", "/api/v1/me",
		"/api/v1/kinds", "/api/v1/categories",
		"/api/v1/items", "/api/v1/items/bulk", "/api/v1/items/export", "/api/v1/items/import",
		"/api/v1/notifications", "/api/v1/events",
		"/api/v1/push/vapid-public", "/api/v1/push/subscribe",
		"/api/v1/dashboard", "/api/v1/calendar", "/api/v1/audit",
	} {
		if !strings.Contains(body, path) {
			t.Fatalf("spec missing %s", path)
		}
	}
}

func TestDocsRedirect(t *testing.T) {
	t.Parallel()
	api := docsAPI()
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/docs", nil)
	api.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status: got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/docs/" {
		t.Fatalf("location: %s", loc)
	}
}

func TestDocsUI(t *testing.T) {
	t.Parallel()
	api := docsAPI()
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/docs/", nil)
	api.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "swagger") && !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("not swagger ui: ct=%s body=%s", rec.Header().Get("Content-Type"), rec.Body.String())
	}
}
