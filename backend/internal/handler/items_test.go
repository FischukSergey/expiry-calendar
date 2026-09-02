package handler_test

import (
	"bytes"
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

const (
	otherKindID     = "33333333-3333-3333-3333-333333333309"
	itemTitleDomain = "Домен"
	fixtureUUID     = "11111111-1111-1111-1111-111111111111"
	kindColorBlack  = "#000"
	kindSlugOther   = "other"
	kindNameOther   = "Прочее"
	expiresSoon     = "2026-09-10"
	currencyUSD     = "USD"
	pathItemsBulk   = "/api/v1/items/bulk"
)

func itemsAPI(t *testing.T) *handler.API {
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

	clk := clock.Fixed{T: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)}
	cats := newMemCats()
	items := service.NewItem(newMemItems(), kinds, cats, newMemRenewals(), newMemAudit(), nopTx, clk)
	return handler.New(handler.Deps{
		Health:        fakeHealth{},
		Auth:          fakeAuth{},
		Kinds:         service.NewKind(kinds),
		Categories:    service.NewCategory(cats),
		Items:         items,
		Notifications: service.NewNotification(newMemNotifications()),
		JWTSecret:     []byte("handler-test-secret"),
		RefreshTTL:    336 * time.Hour,
	})
}

func TestViewerForbiddenCreateItem(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"title":"x","kind_id":"` + otherKindID + `","expires_at":"2027-01-01"}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/items", body)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, string(model.RoleViewer)))
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
}

func TestViewerCanListItems(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items", nil)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, string(model.RoleViewer)))
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
}

func TestAdminCreateGetItem(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	tok := testJWT(t, string(model.RoleAdmin))
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"title":"` + itemTitleDomain + `","kind_id":"` + otherKindID +
		`","expires_at":"2027-01-01","attrs":{}}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/items", body)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	var created model.Item
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.Status == "" || created.Currency != model.CurrencyRUB || created.BillingPeriod != model.BillingOneTime {
		t.Fatalf("defaults %+v", created)
	}
	if created.Status != model.StatusActive {
		t.Fatalf("status %s", created.Status)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items/"+created.ID, nil)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, string(model.RoleViewer)))
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get %d %s", rec.Code, rec.Body.String())
	}
	var card model.ItemCard
	if err := json.NewDecoder(rec.Body).Decode(&card); err != nil {
		t.Fatal(err)
	}
	if card.Item.ID != created.ID || card.Renewals == nil {
		t.Fatalf("card %+v", card)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/audit", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("audit %d %s", rec.Code, rec.Body.String())
	}
	var audit model.AuditList
	if err := json.NewDecoder(rec.Body).Decode(&audit); err != nil {
		t.Fatal(err)
	}
	if audit.Total != 1 || audit.Items[0].Action != model.AuditCreate {
		t.Fatalf("audit %+v", audit)
	}
}

func TestExtraAttrRejected(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"title":"x","kind_id":"` + otherKindID + `","expires_at":"2027-01-01","attrs":{"vin":"1"}}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/items", body)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, string(model.RoleAdmin)))
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
}

func TestViewerForbiddenItemMutations(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	admin := testJWT(t, string(model.RoleAdmin))
	viewer := testJWT(t, string(model.RoleViewer))
	it := adminCreateItem(t, api, admin, `{"title":"x","kind_id":"`+otherKindID+`","expires_at":"2027-01-01"}`)
	for _, tc := range []struct {
		method, path, raw string
	}{
		{http.MethodPatch, "/api/v1/items/" + it.ID, `{"title":"y"}`},
		{http.MethodDelete, "/api/v1/items/" + it.ID, ""},
		{http.MethodPost, "/api/v1/items/" + it.ID + "/renew", `{"new_expires_at":"2028-01-01"}`},
		{http.MethodPost, pathItemsBulk, `{"ids":["` + it.ID + `"],"status":"archived"}`},
		{http.MethodPost, "/api/v1/items/import", ""},
	} {
		rec := httptest.NewRecorder()
		var req *http.Request
		if tc.raw != "" {
			req = httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, bytes.NewBufferString(tc.raw))
		} else {
			req = httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, http.NoBody)
		}
		req.Header.Set("Authorization", "Bearer "+viewer)
		api.Router().ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, rec.Code, rec.Body.String())
		}
	}
}

func TestItemsCRUDRenewFilterPage(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	tok := testJWT(t, string(model.RoleAdmin))
	kind := `","kind_id":"` + otherKindID + `","expires_at":`
	a := adminCreateItem(t, api, tok, `{"title":"Rent-A`+kind+`"2027-01-01","tags":["office"]}`)
	_ = adminCreateItem(t, api, tok, `{"title":"Rent-B`+kind+`"2027-02-01"}`)
	_ = adminCreateItem(t, api, tok, `{"title":"Sub`+kind+`"2027-03-01"}`)

	rec := adminJSON(t, api, tok, http.MethodPatch, "/api/v1/items/"+a.ID, `{"title":"Rent-A2"}`, http.StatusOK)
	var patched model.Item
	if err := json.NewDecoder(rec.Body).Decode(&patched); err != nil || patched.Title != "Rent-A2" {
		t.Fatalf("patch %+v", patched)
	}

	rec = adminJSON(t, api, tok, http.MethodPost, "/api/v1/items/"+a.ID+"/renew",
		`{"new_expires_at":"2028-08-01","new_cost":1990,"comment":"год"}`, http.StatusOK)
	var renewed model.Item
	if err := json.NewDecoder(rec.Body).Decode(&renewed); err != nil {
		t.Fatal(err)
	}
	if renewed.ExpiresAt != "2028-08-01" || renewed.CostAmount != 1990 {
		t.Fatalf("renew %+v", renewed)
	}

	card := getItemCard(t, api, tok, a.ID)
	if len(card.Renewals) != 1 || card.Renewals[0].NewExpiresAt != "2028-08-01" {
		t.Fatalf("renewals %+v", card.Renewals)
	}

	page1 := listItems(t, api, tok, "q=Rent&page=1&per_page=1")
	if page1.Total != 2 || page1.Page != 1 || page1.PerPage != 1 || len(page1.Items) != 1 {
		t.Fatalf("page1 %+v", page1)
	}
	page2 := listItems(t, api, tok, "q=Rent&page=2&per_page=1")
	if page2.Total != 2 || len(page2.Items) != 1 || page1.Items[0].ID == page2.Items[0].ID {
		t.Fatalf("page2 %+v vs page1 %s", page2, page1.Items[0].ID)
	}
	byTag := listItems(t, api, tok, "tag=office")
	if byTag.Total != 1 || byTag.Items[0].ID != a.ID {
		t.Fatalf("tag %+v", byTag)
	}

	adminJSON(t, api, tok, http.MethodDelete, "/api/v1/items/"+a.ID, "", http.StatusNoContent)
	rec = httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items/"+a.ID, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("deleted get %d", rec.Code)
	}
}

func TestForeignOwnerItemMutationsNotFound(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	owner := testJWT(t, string(model.RoleAdmin))
	other := testJWTSub(t, string(model.RoleAdmin), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	it := adminCreateItem(t, api, owner, `{"title":"x","kind_id":"`+otherKindID+`","expires_at":"2027-01-01"}`)
	for _, tc := range []struct {
		method, path, raw string
	}{
		{http.MethodGet, "/api/v1/items/" + it.ID, ""},
		{http.MethodPatch, "/api/v1/items/" + it.ID, `{"title":"y"}`},
		{http.MethodDelete, "/api/v1/items/" + it.ID, ""},
		{http.MethodPost, "/api/v1/items/" + it.ID + "/renew", `{"new_expires_at":"2028-01-01"}`},
		{http.MethodPost, pathItemsBulk, `{"ids":["` + it.ID + `"],"status":"archived"}`},
	} {
		rec := httptest.NewRecorder()
		var req *http.Request
		if tc.raw != "" {
			req = httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, bytes.NewBufferString(tc.raw))
		} else {
			req = httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, http.NoBody)
		}
		req.Header.Set("Authorization", "Bearer "+other)
		api.Router().ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, rec.Code, rec.Body.String())
		}
	}
}

func TestOwnerIsolationItemsListGetAudit(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	owner := testJWT(t, string(model.RoleAdmin))
	other := testJWTSub(t, string(model.RoleAdmin), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	it := adminCreateItem(t, api, owner, `{"title":"x","kind_id":"`+otherKindID+`","expires_at":"2027-01-01"}`)

	theirs := listItems(t, api, other, "")
	if theirs.Total != 0 || len(theirs.Items) != 0 {
		t.Fatalf("leaked list %+v", theirs)
	}
	mine := listItems(t, api, owner, "")
	if mine.Total != 1 || mine.Items[0].ID != it.ID {
		t.Fatalf("own list %+v", mine)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items/"+it.ID, nil)
	req.Header.Set("Authorization", "Bearer "+other)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get %d %s", rec.Code, rec.Body.String())
	}

	var audit model.AuditList
	rec = httptest.NewRecorder()
	req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/audit", nil)
	req.Header.Set("Authorization", "Bearer "+other)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("audit other %d %s", rec.Code, rec.Body.String())
	}
	if err := json.NewDecoder(rec.Body).Decode(&audit); err != nil {
		t.Fatal(err)
	}
	if audit.Total != 0 || len(audit.Items) != 0 {
		t.Fatalf("leaked audit %+v", audit)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/audit", nil)
	req.Header.Set("Authorization", "Bearer "+owner)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("audit own %d %s", rec.Code, rec.Body.String())
	}
	if err := json.NewDecoder(rec.Body).Decode(&audit); err != nil {
		t.Fatal(err)
	}
	if audit.Total != 1 || audit.Items[0].Action != model.AuditCreate {
		t.Fatalf("own audit %+v", audit)
	}
}

func TestForeignCategoryIDNotFound(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	owner := testJWT(t, string(model.RoleAdmin))
	other := testJWTSub(t, string(model.RoleAdmin), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	cat := adminCreateCategory(t, api, owner, `{"name":"Work"}`)

	rec := httptest.NewRecorder()
	raw := `{"title":"x","kind_id":"` + otherKindID + `","expires_at":"2027-01-01","category_id":"` + cat.ID + `"}`
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/items", bytes.NewBufferString(raw))
	req.Header.Set("Authorization", "Bearer "+other)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items?category_id="+cat.ID, nil)
	req.Header.Set("Authorization", "Bearer "+other)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("filter %d %s", rec.Code, rec.Body.String())
	}
}

func TestOwnerIsolationCategories(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	owner := testJWT(t, string(model.RoleAdmin))
	other := testJWTSub(t, string(model.RoleAdmin), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	cat := adminCreateCategory(t, api, owner, `{"name":"Mine"}`)

	own := listCategories(t, api, owner)
	if len(own) != 1 || own[0].ID != cat.ID {
		t.Fatalf("own %+v", own)
	}
	theirs := listCategories(t, api, other)
	if len(theirs) != 0 {
		t.Fatalf("leaked %+v", theirs)
	}

	adminJSON(t, api, other, http.MethodPatch, "/api/v1/categories/"+cat.ID, `{"name":"X"}`, http.StatusNotFound)
	adminJSON(t, api, other, http.MethodDelete, "/api/v1/categories/"+cat.ID, "", http.StatusNotFound)

	rec := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"name":"Child","parent_id":"` + cat.ID + `"}`)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/categories", body)
	req.Header.Set("Authorization", "Bearer "+other)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("parent %d %s", rec.Code, rec.Body.String())
	}
}

func TestKindsRemainShared(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	other := testJWTSub(t, string(model.RoleAdmin), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/kinds", nil)
	req.Header.Set("Authorization", "Bearer "+other)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var out model.KindList
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Items) == 0 {
		t.Fatal("kinds empty")
	}
}

func TestViewerForbiddenAudit(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/audit", nil)
	req.Header.Set("Authorization", "Bearer "+testJWT(t, string(model.RoleViewer)))
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
}

func TestMutationsWriteAudit(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	tok := testJWT(t, string(model.RoleAdmin))
	createBody := `{"title":"Аудит","kind_id":"` + otherKindID + `",` +
		`"expires_at":"2027-01-01","url":"https://secret.example","account_hint":"login"}`
	created := adminCreateItem(t, api, tok, createBody)

	adminJSON(t, api, tok, http.MethodPatch, "/api/v1/items/"+created.ID, `{"title":"Аудит-2"}`, http.StatusOK)
	adminJSON(t, api, tok, http.MethodPost, "/api/v1/items/"+created.ID+"/renew",
		`{"new_expires_at":"2028-01-01","new_cost":10,"comment":"год"}`, http.StatusOK)
	second := adminCreateItem(t, api, tok, `{"title":"Второй","kind_id":"`+otherKindID+`","expires_at":"2027-06-01"}`)
	adminJSON(t, api, tok, http.MethodPost, pathItemsBulk,
		`{"ids":["`+second.ID+`"],"status":"archived"}`, http.StatusOK)
	adminJSON(t, api, tok, http.MethodDelete, "/api/v1/items/"+created.ID, "", http.StatusNoContent)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/audit?per_page=20", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("audit %d %s", rec.Code, rec.Body.String())
	}
	var list model.AuditList
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, e := range list.Items {
		got[e.Action]++
		assertAuditNoSecrets(t, e)
	}
	for _, action := range []string{
		model.AuditCreate, model.AuditUpdate, model.AuditRenew, model.AuditBulk, model.AuditDelete,
	} {
		if got[action] < 1 {
			t.Fatalf("missing action %s in %+v", action, got)
		}
	}
}

func TestCreatePatchPaidAndNullNotify(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	tok := testJWT(t, string(model.RoleAdmin))
	kind := `","kind_id":"` + otherKindID + `","expires_at":`

	paid := adminCreateItem(t, api, tok, `{"title":"Paid`+kind+`"2026-09-10","status":"paid","notify_before_days":30}`)
	if paid.Status != model.StatusPaid || paid.NotifyBeforeDays == nil || *paid.NotifyBeforeDays != 30 {
		t.Fatalf("create paid %+v", paid)
	}
	listed := listItems(t, api, tok, "status=paid")
	if listed.Total != 1 || listed.Items[0].ID != paid.ID {
		t.Fatalf("filter paid %+v", listed)
	}

	rec := adminJSON(t, api, tok, http.MethodPatch, "/api/v1/items/"+paid.ID, `{"notify_before_days":null}`, http.StatusOK)
	var patched model.Item
	if err := json.NewDecoder(rec.Body).Decode(&patched); err != nil {
		t.Fatal(err)
	}
	if patched.NotifyBeforeDays != nil || patched.Status != model.StatusPaid {
		t.Fatalf("patch null notify %+v", patched)
	}

	silent := adminCreateItem(t, api, tok, `{"title":"Quiet`+kind+`"2026-09-05","notify_before_days":null}`)
	if silent.Status != model.StatusActive || silent.NotifyBeforeDays != nil {
		t.Fatalf("null notify %+v", silent)
	}

	def := adminCreateItem(t, api, tok, `{"title":"Def`+kind+`"2027-01-01"}`)
	if def.NotifyBeforeDays == nil || *def.NotifyBeforeDays != model.DefaultNotifyDays {
		t.Fatalf("default notify %+v", def)
	}

	renewed := adminJSON(t, api, tok, http.MethodPost, "/api/v1/items/"+paid.ID+"/renew",
		`{"new_expires_at":"2027-09-10"}`, http.StatusOK)
	var after model.Item
	if err := json.NewDecoder(renewed.Body).Decode(&after); err != nil {
		t.Fatal(err)
	}
	if after.Status == model.StatusPaid {
		t.Fatal("renew kept paid")
	}

	adminJSON(t, api, tok, http.MethodPost, pathItemsBulk,
		`{"ids":["`+silent.ID+`"],"status":"paid"}`, http.StatusOK)
	card := getItemCard(t, api, tok, silent.ID)
	if card.Item.Status != model.StatusPaid {
		t.Fatalf("bulk paid %s", card.Item.Status)
	}
}

func adminCreateItem(t *testing.T, api *handler.API, tok, raw string) model.Item {
	t.Helper()
	rec := adminJSON(t, api, tok, http.MethodPost, "/api/v1/items", raw, http.StatusCreated)
	var it model.Item
	if err := json.NewDecoder(rec.Body).Decode(&it); err != nil {
		t.Fatal(err)
	}
	return it
}

func adminJSON(t *testing.T, api *handler.API, tok, method, path, raw string, want int) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	var body *bytes.Buffer
	if raw != "" {
		body = bytes.NewBufferString(raw)
	}
	var req *http.Request
	if body != nil {
		req = httptest.NewRequestWithContext(t.Context(), method, path, body)
	} else {
		req = httptest.NewRequestWithContext(t.Context(), method, path, http.NoBody)
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != want {
		t.Fatalf("%s %s: %d %s", method, path, rec.Code, rec.Body.String())
	}
	return rec
}

func adminCreateCategory(t *testing.T, api *handler.API, tok, raw string) model.Category {
	t.Helper()
	rec := adminJSON(t, api, tok, http.MethodPost, "/api/v1/categories", raw, http.StatusCreated)
	var c model.Category
	if err := json.NewDecoder(rec.Body).Decode(&c); err != nil {
		t.Fatal(err)
	}
	return c
}

func listCategories(t *testing.T, api *handler.API, tok string) []model.Category {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/categories", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("categories %d %s", rec.Code, rec.Body.String())
	}
	var out model.CategoryList
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out.Items
}

func listItems(t *testing.T, api *handler.API, tok, query string) model.ItemList {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items?"+query, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list %d %s", rec.Code, rec.Body.String())
	}
	var out model.ItemList
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}

func getItemCard(t *testing.T, api *handler.API, tok, id string) model.ItemCard {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get %d %s", rec.Code, rec.Body.String())
	}
	var card model.ItemCard
	if err := json.NewDecoder(rec.Body).Decode(&card); err != nil {
		t.Fatal(err)
	}
	return card
}

func assertAuditNoSecrets(t *testing.T, e model.AuditEntry) {
	t.Helper()
	for _, raw := range []json.RawMessage{e.BeforeJSON, e.AfterJSON} {
		if len(raw) == 0 {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		for _, banned := range []string{"password", "password_hash", "account_hint", "url"} {
			if _, ok := m[banned]; ok {
				t.Fatalf("audit %s has %s: %s", e.Action, banned, raw)
			}
		}
	}
}
