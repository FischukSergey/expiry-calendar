package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"duekeep/internal/model"
)

func testJWT(t *testing.T, role string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  fixtureUUID,
		"role": role,
		"iss":  model.JWTIssuer,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	s, err := tok.SignedString([]byte("handler-test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestKindsRequiresAuth(t *testing.T) {
	t.Parallel()
	api := testAPI(fakeAuth{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/kinds", nil)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestViewerForbiddenCreateKind(t *testing.T) {
	t.Parallel()
	api := testAPI(fakeAuth{})
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"slug":"visa","name":"Виза","color":"#000"}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/kinds", body)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, string(model.RoleViewer)))
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
}

func TestViewerCanListKinds(t *testing.T) {
	t.Parallel()
	api := testAPI(fakeAuth{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/kinds", nil)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, string(model.RoleViewer)))
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
}
