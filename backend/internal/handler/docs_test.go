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
	return handler.New(handler.Deps{Health: fakeHealth{}, Auth: fakeAuth{}, Spec: duekeep.OpenAPISpec, JWTSecret: []byte("x")})
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
