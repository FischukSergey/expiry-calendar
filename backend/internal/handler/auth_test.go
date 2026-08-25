package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"duekeep/internal/handler"
	"duekeep/internal/model"
)

type fakeAuth struct {
	pair model.TokenPair
	me   model.PublicUser
	err  error
}

func (f fakeAuth) Register(context.Context, string, string, string) (model.TokenPair, error) {
	return f.pair, f.err
}

func (f fakeAuth) Login(context.Context, string, string, string) (model.TokenPair, error) {
	return f.pair, f.err
}

func (f fakeAuth) Refresh(context.Context, string, string) (model.TokenPair, error) {
	return f.pair, f.err
}

func (f fakeAuth) Logout(context.Context, string) error { return f.err }

func (f fakeAuth) LogoutAll(context.Context, string) error { return f.err }

func (f fakeAuth) Me(context.Context, string) (model.PublicUser, error) {
	return f.me, f.err
}

func testAPI(auth handler.AuthService) *handler.API {
	return handler.New(handler.Deps{
		Health:     fakeHealth{},
		Auth:       auth,
		JWTSecret:  []byte("handler-test-secret"),
		RefreshTTL: 336 * time.Hour,
	})
}

func TestLoginSetsCookie(t *testing.T) {
	t.Parallel()
	api := testAPI(fakeAuth{pair: model.TokenPair{
		AccessToken:  "acc",
		RefreshToken: "ref-raw",
		TokenType:    model.TokenTypeBearer,
		ExpiresIn:    900,
	}})
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"email":"admin@duekeep.local","password":"admin1234"}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/login", body)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	cookie := rec.Result().Cookies()
	found := false
	for _, c := range cookie {
		if c.Name == model.RefreshCookie && c.Value == "ref-raw" && c.Path == model.RefreshCookiePath && c.HttpOnly {
			found = true
		}
	}
	if !found {
		t.Fatalf("cookie: %+v", cookie)
	}
}

func TestMeUnauthorizedWithoutBearer(t *testing.T) {
	t.Parallel()
	api := testAPI(fakeAuth{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/me", nil)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "unauthorized") {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestLogoutRequiresAccessOrRefresh(t *testing.T) {
	t.Parallel()
	api := testAPI(fakeAuth{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/logout", nil)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestRefreshPrefersBodyOverCookie(t *testing.T) {
	t.Parallel()
	var gotRaw string
	api := testAPI(captureRefresh{out: &gotRaw, pair: model.TokenPair{
		AccessToken: "a", RefreshToken: "new", TokenType: "Bearer", ExpiresIn: 900,
	}})
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"refresh_token":"from-body"}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/refresh", body)
	req.AddCookie(&http.Cookie{Name: model.RefreshCookie, Value: "from-cookie"}) //nolint:gosec // G124: тестовый входящий cookie.
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	if gotRaw != "from-body" {
		t.Fatalf("raw %q", gotRaw)
	}
}

func TestRegisterJSON(t *testing.T) {
	t.Parallel()
	api := testAPI(fakeAuth{pair: model.TokenPair{
		AccessToken: "a", RefreshToken: "r", TokenType: "Bearer", ExpiresIn: 900,
	}})
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"email":"new@duekeep.local","password":"secret12"}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/register", body)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d", rec.Code)
	}
	var pair model.TokenPair
	if err := json.NewDecoder(rec.Body).Decode(&pair); err != nil {
		t.Fatal(err)
	}
	if pair.RefreshToken != "r" {
		t.Fatalf("%+v", pair)
	}
}

type captureRefresh struct {
	out  *string
	pair model.TokenPair
}

func (c captureRefresh) Register(context.Context, string, string, string) (model.TokenPair, error) {
	return model.TokenPair{}, nil
}

func (c captureRefresh) Login(context.Context, string, string, string) (model.TokenPair, error) {
	return model.TokenPair{}, nil
}

func (c captureRefresh) Refresh(_ context.Context, raw, _ string) (model.TokenPair, error) {
	*c.out = raw
	return c.pair, nil
}

func (c captureRefresh) Logout(context.Context, string) error { return nil }

func (c captureRefresh) LogoutAll(context.Context, string) error { return nil }

func (c captureRefresh) Me(context.Context, string) (model.PublicUser, error) {
	return model.PublicUser{}, nil
}
