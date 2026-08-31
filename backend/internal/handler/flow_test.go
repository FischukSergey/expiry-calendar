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
	"duekeep/internal/seed"
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

func liveCatalogAPI(t *testing.T) (*handler.API, *memItems, *memCats) {
	t.Helper()
	users := newMemUsers()
	cats := newMemCats()
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
	auth := service.NewAuth(users, newMemRefresh(), nopTx, clock.Real{}, service.AuthConfig{
		Secret:     []byte("handler-test-secret"),
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 336 * time.Hour,
		BcryptCost: 4,
	})
	auth.SetCategoryDefaults(cats)
	clk := clock.Fixed{T: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)}
	api := handler.New(handler.Deps{
		Health:        fakeHealth{},
		Auth:          auth,
		Kinds:         service.NewKind(kinds),
		Categories:    service.NewCategory(cats),
		Items:         service.NewItem(store, kinds, cats, newMemRenewals(), newMemAudit(), nopTx, clk),
		Overview:      service.NewOverview(store, clk),
		Notifications: nopNotifications{},
		JWTSecret:     []byte("handler-test-secret"),
		RefreshTTL:    336 * time.Hour,
	})
	return api, store, cats
}

func TestRegisterDoesNotSeeSeedCatalog(t *testing.T) {
	t.Parallel()
	api, store, cats := liveCatalogAPI(t)
	seedCat, err := cats.Create(t.Context(), model.Category{OwnerID: fixtureUUID, Name: "Seed IT"})
	if err != nil {
		t.Fatal(err)
	}
	seedItem, err := store.Create(t.Context(), model.Item{
		OwnerID: fixtureUUID, Title: "duekeep.ru", KindID: otherKindID,
		Status: model.StatusActive, ExpiresAt: "2027-01-01",
		Tags: []string{}, Attrs: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}

	regA := serveJSON(t, api, http.MethodPost, "/api/v1/auth/register",
		`{"email":"a@duekeep.local","password":"secret12"}`, "")
	if regA.Code != http.StatusCreated {
		t.Fatalf("register A %d %s", regA.Code, regA.Body.String())
	}
	pairA := decodePair(t, regA)
	idA, err := service.ParseAccess([]byte("handler-test-secret"), pairA.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if idA.Role != string(model.RoleAdmin) {
		t.Fatalf("role A %s", idA.Role)
	}

	me := serveJSON(t, api, http.MethodGet, "/api/v1/me", "", pairA.AccessToken)
	if me.Code != http.StatusOK {
		t.Fatalf("me %d %s", me.Code, me.Body.String())
	}
	var pub model.PublicUser
	if err := json.NewDecoder(me.Body).Decode(&pub); err != nil {
		t.Fatal(err)
	}
	if pub.Email != "a@duekeep.local" || pub.Role != model.RoleAdmin || pub.ID == fixtureUUID {
		t.Fatalf("me %+v", pub)
	}

	if got := listItems(t, api, pairA.AccessToken, ""); got.Total != 0 || len(got.Items) != 0 {
		t.Fatalf("A saw seed items %+v", got)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items/"+seedItem.ID, nil)
	req.Header.Set("Authorization", "Bearer "+pairA.AccessToken)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("seed get %d %s", rec.Code, rec.Body.String())
	}

	dash := serveJSON(t, api, http.MethodGet, "/api/v1/dashboard", "", pairA.AccessToken)
	if dash.Code != http.StatusOK {
		t.Fatalf("dashboard %d %s", dash.Code, dash.Body.String())
	}
	var board model.Dashboard
	if err := json.NewDecoder(dash.Body).Decode(&board); err != nil {
		t.Fatal(err)
	}
	if board.Counts.Active != 0 || len(board.Soonest) != 0 {
		t.Fatalf("dashboard leaked seed %+v", board)
	}

	tree := listCategories(t, api, pairA.AccessToken)
	if n := countCatNodes(tree); n != len(seed.DefaultCategories()) {
		t.Fatalf("A cats %d", n)
	}
	var walk func([]model.Category)
	walk = func(rows []model.Category) {
		for _, c := range rows {
			if c.ID == seedCat.ID {
				t.Fatal("A saw seed category")
			}
			walk(c.Children)
		}
	}
	walk(tree)

	created := adminCreateItem(t, api, pairA.AccessToken,
		`{"title":"Мой домен","kind_id":"`+otherKindID+`","expires_at":"2027-06-01"}`)

	regB := serveJSON(t, api, http.MethodPost, "/api/v1/auth/register",
		`{"email":"b@duekeep.local","password":"secret12"}`, "")
	if regB.Code != http.StatusCreated {
		t.Fatalf("register B %d %s", regB.Code, regB.Body.String())
	}
	pairB := decodePair(t, regB)
	if got := listItems(t, api, pairB.AccessToken, ""); got.Total != 0 || len(got.Items) != 0 {
		t.Fatalf("B saw items %+v", got)
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items/"+created.ID, nil)
	req.Header.Set("Authorization", "Bearer "+pairB.AccessToken)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("B get A %d %s", rec.Code, rec.Body.String())
	}
}

func TestRegisterCopiesDefaultCategoriesNoItems(t *testing.T) {
	t.Parallel()
	users := newMemUsers()
	cats := newMemCats()
	kinds := newMemKinds()
	auth := service.NewAuth(users, newMemRefresh(), nopTx, clock.Real{}, service.AuthConfig{
		Secret:     []byte("handler-test-secret"),
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 336 * time.Hour,
		BcryptCost: 4,
	})
	auth.SetCategoryDefaults(cats)
	clk := clock.Real{}
	api := handler.New(handler.Deps{
		Health:        fakeHealth{},
		Auth:          auth,
		Kinds:         service.NewKind(kinds),
		Categories:    service.NewCategory(cats),
		Items:         service.NewItem(newMemItems(), kinds, cats, newMemRenewals(), newMemAudit(), nopTx, clk),
		Notifications: nopNotifications{},
		JWTSecret:     []byte("handler-test-secret"),
		RefreshTTL:    336 * time.Hour,
	})

	reg := serveJSON(t, api, http.MethodPost, "/api/v1/auth/register",
		`{"email":"own@duekeep.local","password":"secret12"}`, "")
	if reg.Code != http.StatusCreated {
		t.Fatalf("register %d %s", reg.Code, reg.Body.String())
	}
	pair := decodePair(t, reg)

	tree := listCategories(t, api, pair.AccessToken)
	if n := countCatNodes(tree); n != len(seed.DefaultCategories()) {
		t.Fatalf("categories %d want %d", n, len(seed.DefaultCategories()))
	}
	items := listItems(t, api, pair.AccessToken, "")
	if items.Total != 0 || len(items.Items) != 0 {
		t.Fatalf("items %+v", items)
	}
}

func countCatNodes(tree []model.Category) int {
	n := 0
	var walk func([]model.Category)
	walk = func(rows []model.Category) {
		for _, c := range rows {
			n++
			walk(c.Children)
		}
	}
	walk(tree)
	return n
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
