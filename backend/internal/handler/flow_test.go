package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"duekeep/internal/clock"
	"duekeep/internal/handler"
	"duekeep/internal/model"
	"duekeep/internal/service"
)

func nopTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func liveAPI(t *testing.T, users *memUsers) *handler.API {
	t.Helper()
	if users == nil {
		users = newMemUsers()
	}
	auth := service.NewAuth(users, newMemRefresh(), nopTx, clock.Real{}, service.AuthConfig{
		Secret:     []byte("handler-test-secret"),
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 336 * time.Hour,
		BcryptCost: 4,
	})
	return handler.New(handler.Deps{
		Health:        fakeHealth{},
		Auth:          auth,
		Kinds:         service.NewKind(newMemKinds()),
		Categories:    nopCategories{},
		Items:         nopItems{},
		Notifications: nopNotifications{},
		JWTSecret:     []byte("handler-test-secret"),
		RefreshTTL:    336 * time.Hour,
	})
}

func serveJSON(t *testing.T, api *handler.API, method, path, raw, access string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	var body *bytes.Buffer
	if raw != "" {
		body = bytes.NewBufferString(raw)
	} else {
		body = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequestWithContext(t.Context(), method, path, body)
	if raw != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if access != "" {
		req.Header.Set("Authorization", "Bearer "+access)
	}
	api.Router().ServeHTTP(rec, req)
	return rec
}

func decodePair(t *testing.T, rec *httptest.ResponseRecorder) model.TokenPair {
	t.Helper()
	var pair model.TokenPair
	if err := json.NewDecoder(rec.Body).Decode(&pair); err != nil {
		t.Fatalf("decode pair: %v body %s", err, rec.Body.String())
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" || pair.TokenType != model.TokenTypeBearer {
		t.Fatalf("pair %+v", pair)
	}
	return pair
}

func TestLoginRefreshLogout(t *testing.T) {
	t.Parallel()
	api := liveAPI(t, nil)
	reg := serveJSON(t, api, http.MethodPost, "/api/v1/auth/register",
		`{"email":"flow@duekeep.local","password":"secret12"}`, "")
	if reg.Code != http.StatusCreated {
		t.Fatalf("register %d %s", reg.Code, reg.Body.String())
	}
	first := decodePair(t, reg)
	if !hasRefreshCookie(reg, first.RefreshToken) {
		t.Fatal("register cookie")
	}

	ref := serveJSON(t, api, http.MethodPost, "/api/v1/auth/refresh",
		`{"refresh_token":"`+first.RefreshToken+`"}`, "")
	if ref.Code != http.StatusOK {
		t.Fatalf("refresh %d %s", ref.Code, ref.Body.String())
	}
	second := decodePair(t, ref)
	if second.RefreshToken == first.RefreshToken {
		t.Fatal("refresh must rotate")
	}

	out := serveJSON(t, api, http.MethodPost, "/api/v1/auth/logout",
		`{"refresh_token":"`+second.RefreshToken+`"}`, second.AccessToken)
	if out.Code != http.StatusNoContent {
		t.Fatalf("logout %d %s", out.Code, out.Body.String())
	}

	after := serveJSON(t, api, http.MethodPost, "/api/v1/auth/refresh",
		`{"refresh_token":"`+second.RefreshToken+`"}`, "")
	if after.Code != http.StatusUnauthorized {
		t.Fatalf("refresh after logout %d %s", after.Code, after.Body.String())
	}
}

func TestRefreshReuseRevokesFamily(t *testing.T) {
	t.Parallel()
	api := liveAPI(t, nil)
	reg := serveJSON(t, api, http.MethodPost, "/api/v1/auth/register",
		`{"email":"reuse@duekeep.local","password":"secret12"}`, "")
	if reg.Code != http.StatusCreated {
		t.Fatalf("register %d %s", reg.Code, reg.Body.String())
	}
	first := decodePair(t, reg)
	ref := serveJSON(t, api, http.MethodPost, "/api/v1/auth/refresh",
		`{"refresh_token":"`+first.RefreshToken+`"}`, "")
	if ref.Code != http.StatusOK {
		t.Fatalf("refresh %d %s", ref.Code, ref.Body.String())
	}
	second := decodePair(t, ref)

	old := serveJSON(t, api, http.MethodPost, "/api/v1/auth/refresh",
		`{"refresh_token":"`+first.RefreshToken+`"}`, "")
	if old.Code != http.StatusUnauthorized {
		t.Fatalf("reuse %d %s", old.Code, old.Body.String())
	}
	dead := serveJSON(t, api, http.MethodPost, "/api/v1/auth/refresh",
		`{"refresh_token":"`+second.RefreshToken+`"}`, "")
	if dead.Code != http.StatusUnauthorized {
		t.Fatalf("family %d %s", dead.Code, dead.Body.String())
	}
}

func TestViewerForbiddenCreateKindLive(t *testing.T) {
	t.Parallel()
	users := newMemUsers()
	hash, err := bcrypt.GenerateFromPassword([]byte("secret12"), 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := users.Create(t.Context(), "viewer@duekeep.local", string(hash), model.RoleViewer); err != nil {
		t.Fatal(err)
	}
	api := liveAPI(t, users)
	login := serveJSON(t, api, http.MethodPost, "/api/v1/auth/login",
		`{"email":"viewer@duekeep.local","password":"secret12"}`, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login %d %s", login.Code, login.Body.String())
	}
	pair := decodePair(t, login)

	list := serveJSON(t, api, http.MethodGet, "/api/v1/kinds", "", pair.AccessToken)
	if list.Code != http.StatusOK {
		t.Fatalf("list %d %s", list.Code, list.Body.String())
	}
	create := serveJSON(t, api, http.MethodPost, "/api/v1/kinds",
		`{"slug":"visa","name":"Виза","color":"#111111"}`, pair.AccessToken)
	if create.Code != http.StatusForbidden {
		t.Fatalf("create %d %s", create.Code, create.Body.String())
	}
}

func TestRegisterCreatesAdminCanWriteKinds(t *testing.T) {
	t.Parallel()
	api := liveAPI(t, nil)
	reg := serveJSON(t, api, http.MethodPost, "/api/v1/auth/register",
		`{"email":"newadmin@duekeep.local","password":"secret12"}`, "")
	if reg.Code != http.StatusCreated {
		t.Fatalf("register %d %s", reg.Code, reg.Body.String())
	}
	pair := decodePair(t, reg)
	create := serveJSON(t, api, http.MethodPost, "/api/v1/kinds",
		`{"slug":"visa","name":"Виза","color":"#111111"}`, pair.AccessToken)
	if create.Code != http.StatusCreated {
		t.Fatalf("create %d %s", create.Code, create.Body.String())
	}
}

func TestAdminLoginCreatesKind(t *testing.T) {
	t.Parallel()
	users := newMemUsers()
	hash, err := bcrypt.GenerateFromPassword([]byte("admin1234"), 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := users.Create(t.Context(), "admin@duekeep.local", string(hash), model.RoleAdmin); err != nil {
		t.Fatal(err)
	}
	api := liveAPI(t, users)

	login := serveJSON(t, api, http.MethodPost, "/api/v1/auth/login",
		`{"email":"admin@duekeep.local","password":"admin1234"}`, "")
	if login.Code != http.StatusOK {
		t.Fatalf("login %d %s", login.Code, login.Body.String())
	}
	pair := decodePair(t, login)

	me := serveJSON(t, api, http.MethodGet, "/api/v1/me", "", pair.AccessToken)
	if me.Code != http.StatusOK {
		t.Fatalf("me %d %s", me.Code, me.Body.String())
	}
	var pub model.PublicUser
	if err := json.NewDecoder(me.Body).Decode(&pub); err != nil {
		t.Fatal(err)
	}
	if pub.Role != model.RoleAdmin || pub.Email != "admin@duekeep.local" {
		t.Fatalf("me %+v", pub)
	}

	create := serveJSON(t, api, http.MethodPost, "/api/v1/kinds",
		`{"slug":"visa","name":"Виза","color":"#111111"}`, pair.AccessToken)
	if create.Code != http.StatusCreated {
		t.Fatalf("create %d %s", create.Code, create.Body.String())
	}
}

func hasRefreshCookie(rec *httptest.ResponseRecorder, raw string) bool {
	for _, c := range rec.Result().Cookies() {
		if c.Name == model.RefreshCookie && c.Value == raw &&
			c.Path == model.RefreshCookiePath && c.HttpOnly {
			return true
		}
	}
	return false
}
