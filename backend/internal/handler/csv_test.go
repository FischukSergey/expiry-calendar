package handler_test

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"duekeep/internal/clock"
	"duekeep/internal/handler"
	"duekeep/internal/model"
	"duekeep/internal/service"
)

const (
	domainKindID   = "33333333-3333-3333-3333-333333333310"
	kindSlugDomain = "domain"
	attrRegistrar  = "registrar"
	catWorkName    = "Работа"
)

func csvAPI(t *testing.T) (*handler.API, *memAudit) {
	t.Helper()
	kinds := newMemKinds()
	_, err := kinds.Create(t.Context(), model.Kind{
		Slug: kindSlugDomain, Name: "Домен", Color: kindColorBlack,
		AttrSchema: []model.AttrField{{Key: attrRegistrar, Label: "Рег", Type: model.AttrString}},
	})
	if err != nil {
		t.Fatal(err)
	}
	kinds.mu.Lock()
	k := kinds.byID[fixtureUUID]
	k.ID = domainKindID
	delete(kinds.byID, fixtureUUID)
	kinds.byID[domainKindID] = k
	kinds.mu.Unlock()

	cats := newMemCats()
	if _, err := cats.Create(t.Context(), model.Category{OwnerID: fixtureUUID, Name: catWorkName}); err != nil {
		t.Fatal(err)
	}
	audit := newMemAudit()
	clk := clock.Fixed{T: time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)}
	items := service.NewItem(newMemItems(), kinds, cats, newMemRenewals(), audit, nopTx, clk)
	api := handler.New(handler.Deps{
		Health:        fakeHealth{},
		Auth:          fakeAuth{},
		Kinds:         service.NewKind(kinds),
		Categories:    service.NewCategory(cats),
		Items:         items,
		Notifications: service.NewNotification(newMemNotifications()),
		JWTSecret:     []byte("handler-test-secret"),
		RefreshTTL:    336 * time.Hour,
	})
	return api, audit
}

func TestViewerCanExportForbiddenImport(t *testing.T) {
	t.Parallel()
	api, _ := csvAPI(t)
	viewer := testJWT(t, string(model.RoleViewer))

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items/export", nil)
	req.Header.Set("Authorization", "Bearer "+viewer)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("export %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("ct %s", rec.Header().Get("Content-Type"))
	}

	rec = postCSV(t, api, viewer, "/api/v1/items/import?dry_run=true",
		"title,kind_slug,expires_at\nA,domain,2027-01-01\n",
		`{"title":"title","kind_slug":"kind_slug","expires_at":"expires_at"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("import %d %s", rec.Code, rec.Body.String())
	}
}

func TestExportFilterAndAttrs(t *testing.T) {
	t.Parallel()
	api, _ := csvAPI(t)
	tok := testJWT(t, string(model.RoleAdmin))
	keep := adminCreateItem(t, api, tok, `{"title":"KeepMe","kind_id":"`+domainKindID+
		`","expires_at":"2027-01-01","attrs":{"`+attrRegistrar+`":"reg.ru"}}`)
	_ = adminCreateItem(t, api, tok, `{"title":"Skip","kind_id":"`+domainKindID+`","expires_at":"2027-02-01","attrs":{}}`)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items/export?q=KeepMe", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("export %d %s", rec.Code, rec.Body.String())
	}
	r := csv.NewReader(rec.Body)
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows %d %+v", len(rows), rows)
	}
	header := rows[0]
	if header[0] != "id" || header[2] != "kind_slug" {
		t.Fatalf("header %+v", header)
	}
	attrCol := -1
	for i, h := range header {
		if h == "attrs."+attrRegistrar {
			attrCol = i
		}
	}
	if attrCol < 0 {
		t.Fatalf("no attrs column in %+v", header)
	}
	if rows[1][0] != keep.ID || rows[1][2] != kindSlugDomain || rows[1][attrCol] != "reg.ru" {
		t.Fatalf("row %+v", rows[1])
	}
}

func TestExportOwnerIsolation(t *testing.T) {
	t.Parallel()
	api, _ := csvAPI(t)
	owner := testJWT(t, string(model.RoleAdmin))
	other := testJWTSub(t, string(model.RoleAdmin), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	_ = adminCreateItem(t, api, owner, `{"title":"KeepMe","kind_id":"`+domainKindID+
		`","expires_at":"2027-01-01","attrs":{}}`)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/items/export", nil)
	req.Header.Set("Authorization", "Bearer "+other)
	api.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("export %d %s", rec.Code, rec.Body.String())
	}
	r := csv.NewReader(rec.Body)
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("leaked rows %+v", rows)
	}
}

func TestImportDryRunDoesNotWrite(t *testing.T) {
	t.Parallel()
	api, audit := csvAPI(t)
	tok := testJWT(t, string(model.RoleAdmin))
	rec := postCSV(t, api, tok, "/api/v1/items/import?dry_run=true",
		"Name,Type,Until\nOk,domain,2027-01-01\nBad,domain,\n",
		`{"title":"Name","kind_slug":"Type","expires_at":"Until"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var preview model.CSVImportPreview
	if err := json.NewDecoder(rec.Body).Decode(&preview); err != nil {
		t.Fatal(err)
	}
	if preview.Rows != 2 || preview.Valid != 1 || len(preview.Errors) != 1 {
		t.Fatalf("%+v", preview)
	}
	if preview.Errors[0].Line != 3 {
		t.Fatalf("line %d", preview.Errors[0].Line)
	}
	list := listItems(t, api, tok, "")
	if list.Total != 0 {
		t.Fatalf("wrote %d", list.Total)
	}
	if len(audit.rows) != 0 {
		t.Fatalf("audit %+v", audit.rows)
	}
}

func TestImportWritesBatchAndAudit(t *testing.T) {
	t.Parallel()
	api, audit := csvAPI(t)
	tok := testJWT(t, string(model.RoleAdmin))
	body := "Name,Type,Until,Reg,Cat,Tags\n" +
		"Site,domain,2027-06-01,reg.ru," + catWorkName + ",\"web,dns\"\n"
	rec := postCSV(t, api, tok, "/api/v1/items/import", body,
		`{"title":"Name","kind_slug":"Type","expires_at":"Until","attrs.registrar":"Reg","category_name":"Cat","tags":"Tags"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	var out model.CSVImportResult
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Created != 1 {
		t.Fatalf("%+v", out)
	}
	list := listItems(t, api, tok, "")
	if list.Total != 1 || list.Items[0].Title != "Site" {
		t.Fatalf("list %+v", list)
	}
	if list.Items[0].Attrs[attrRegistrar] != "reg.ru" {
		t.Fatalf("attrs %+v", list.Items[0].Attrs)
	}
	if list.Items[0].CategoryID == nil {
		t.Fatal("category")
	}
	if len(audit.rows) != 1 || audit.rows[0].Action != model.AuditImport {
		t.Fatalf("audit %+v", audit.rows)
	}
	assertAuditNoSecrets(t, audit.rows[0])
}

func TestImportValidationWritesNothing(t *testing.T) {
	t.Parallel()
	api, _ := csvAPI(t)
	tok := testJWT(t, string(model.RoleAdmin))
	rec := postCSV(t, api, tok, "/api/v1/items/import",
		"Name,Type,Until\nOk,domain,2027-01-01\nBad,nope,2027-01-01\n",
		`{"title":"Name","kind_slug":"Type","expires_at":"Until"}`)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	list := listItems(t, api, tok, "")
	if list.Total != 0 {
		t.Fatalf("wrote %d", list.Total)
	}
}

func postCSV(t *testing.T, api *handler.API, tok, path, csvBody, mapping string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", "items.csv")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(part, csvBody); err != nil {
		t.Fatal(err)
	}
	if err := mw.WriteField("mapping", mapping); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, &buf)
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	api.Router().ServeHTTP(rec, req)
	return rec
}
