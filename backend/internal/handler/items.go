package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"duekeep/internal/middleware"
	"duekeep/internal/model"
	"duekeep/internal/service"
)

const (
	maxCSVUpload = 2 << 20
	detailInt    = "int"
)

type itemWrite struct {
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	KindID           string          `json:"kind_id"`
	CategoryID       *string         `json:"category_id"`
	Vendor           string          `json:"vendor"`
	Tags             []string        `json:"tags"`
	CostAmount       *int            `json:"cost_amount"`
	Currency         string          `json:"currency"`
	BillingPeriod    string          `json:"billing_period"`
	StartedAt        *string         `json:"started_at"`
	ExpiresAt        string          `json:"expires_at"`
	NotifyBeforeDays json.RawMessage `json:"notify_before_days"`
	URL              string          `json:"url"`
	AccountHint      string          `json:"account_hint"`
	Status           string          `json:"status"`
	Attrs            map[string]any  `json:"attrs"`
}

type itemPatchBody struct {
	Title            *string         `json:"title"`
	Description      *string         `json:"description"`
	KindID           *string         `json:"kind_id"`
	CategoryID       json.RawMessage `json:"category_id"`
	Vendor           *string         `json:"vendor"`
	Tags             *[]string       `json:"tags"`
	CostAmount       *int            `json:"cost_amount"`
	Currency         *string         `json:"currency"`
	BillingPeriod    *string         `json:"billing_period"`
	StartedAt        json.RawMessage `json:"started_at"`
	ExpiresAt        *string         `json:"expires_at"`
	NotifyBeforeDays json.RawMessage `json:"notify_before_days"`
	URL              *string         `json:"url"`
	AccountHint      *string         `json:"account_hint"`
	Status           *string         `json:"status"`
	Attrs            json.RawMessage `json:"attrs"`
}

type renewBody struct {
	NewExpiresAt string `json:"new_expires_at"`
	NewCost      *int   `json:"new_cost"`
	Comment      string `json:"comment"`
}

type bulkBody struct {
	IDs        []string `json:"ids"`
	CategoryID *string  `json:"category_id"`
	Status     *string  `json:"status"`
}

func (a *API) listItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := queryPage(q.Get("page"), q.Get("per_page"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	f, err := itemFilterFromQuery(q)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out, err := a.items.List(r.Context(), f, page, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, out)
}

func (a *API) exportItems(w http.ResponseWriter, r *http.Request) {
	f, err := itemFilterFromQuery(r.URL.Query())
	if err != nil {
		writeDomainError(w, err)
		return
	}
	body, err := a.items.Export(r.Context(), f, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="items.csv"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body) //nolint:gosec // G705: text/csv, не HTML.
}

func (a *API) importItems(w http.ResponseWriter, r *http.Request) {
	dryRun, err := parseDryRun(r.URL.Query().Get("dry_run"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxCSVUpload)
	if err := r.ParseMultipartForm(maxCSVUpload); err != nil { //nolint:gosec // G120: тело уже ограничено MaxBytesReader.
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid multipart")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "file required")
		return
	}
	defer func() { _ = file.Close() }()
	csvData, err := io.ReadAll(io.LimitReader(file, maxCSVUpload+1))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid file")
		return
	}
	if len(csvData) > maxCSVUpload {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "file too large")
		return
	}
	var mapping map[string]string
	rawMap := r.FormValue("mapping")
	if rawMap == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "mapping required")
		return
	}
	if err := json.Unmarshal([]byte(rawMap), &mapping); err != nil || mapping == nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid mapping")
		return
	}
	preview, created, err := a.items.Import(r.Context(), csvData, mapping, dryRun, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if dryRun {
		writeBytes(w, http.StatusOK, preview)
		return
	}
	writeBytes(w, http.StatusOK, created)
}

func (a *API) createItem(w http.ResponseWriter, r *http.Request) {
	var body itemWrite
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	in, err := itemFromWrite(body)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid notify_before_days")
		return
	}
	it, err := a.items.Create(r.Context(), in, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusCreated, it)
}

func (a *API) getItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	card, err := a.items.Get(r.Context(), id, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, card)
}

func (a *API) patchItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body itemPatchBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	p, err := itemPatchFromBody(body)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}
	it, err := a.items.Patch(r.Context(), id, p, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, it)
}

func (a *API) deleteItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := a.items.Delete(r.Context(), id, middleware.UserID(r.Context())); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type payBody struct {
	Date string `json:"date"`
}

func (a *API) payItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body payBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	out, created, err := a.items.Pay(r.Context(), id, body.Date, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	if created {
		writeBytes(w, http.StatusCreated, out)
		return
	}
	writeBytes(w, http.StatusOK, out)
}

func (a *API) unpayItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	date := r.URL.Query().Get("date")
	if date == "" {
		writeErrorDetails(w, http.StatusUnprocessableEntity, "validation_error", "date required", map[string]any{"date": "required"})
		return
	}
	if err := a.items.Unpay(r.Context(), id, date, middleware.UserID(r.Context())); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) renewItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body renewBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	it, err := a.items.Renew(r.Context(), id, model.RenewInput{
		NewExpiresAt: body.NewExpiresAt, NewCost: body.NewCost, Comment: body.Comment,
	}, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, it)
}

func (a *API) bulkItems(w http.ResponseWriter, r *http.Request) {
	var body bulkBody
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	out, err := a.items.Bulk(r.Context(), model.BulkInput{
		IDs: body.IDs, CategoryID: body.CategoryID, Status: body.Status,
	}, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, out)
}

func (a *API) listAudit(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, err := queryPage(q.Get("page"), q.Get("per_page"))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	out, err := a.items.ListAudit(r.Context(), page, middleware.UserID(r.Context()))
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, out)
}

func itemFromWrite(body itemWrite) (model.Item, error) {
	cost := 0
	if body.CostAmount != nil {
		cost = *body.CostAmount
	}
	notify, set, err := parseOptionalInt(body.NotifyBeforeDays)
	if err != nil {
		return model.Item{}, err
	}
	if !set {
		notify = model.Ptr(model.DefaultNotifyDays)
	}
	tags := body.Tags
	if tags == nil {
		tags = []string{}
	}
	attrs := body.Attrs
	if attrs == nil {
		attrs = map[string]any{}
	}
	return model.Item{
		Title: body.Title, Description: body.Description, KindID: body.KindID,
		CategoryID: body.CategoryID, Vendor: body.Vendor, Tags: tags,
		CostAmount: cost, Currency: body.Currency, BillingPeriod: body.BillingPeriod,
		StartedAt: body.StartedAt, ExpiresAt: body.ExpiresAt, NotifyBeforeDays: notify,
		URL: body.URL, AccountHint: body.AccountHint, Status: body.Status, Attrs: attrs,
	}, nil
}

func itemPatchFromBody(body itemPatchBody) (model.ItemPatch, error) {
	p := model.ItemPatch{
		Title: body.Title, Description: body.Description, KindID: body.KindID,
		Vendor: body.Vendor, Tags: body.Tags, CostAmount: body.CostAmount,
		Currency: body.Currency, BillingPeriod: body.BillingPeriod, ExpiresAt: body.ExpiresAt,
		URL: body.URL, AccountHint: body.AccountHint, Status: body.Status,
	}
	notify, setNotify, err := parseOptionalInt(body.NotifyBeforeDays)
	if err != nil {
		return p, err
	}
	if setNotify {
		p.SetNotify = true
		p.NotifyBeforeDays = notify
	}
	if body.CategoryID != nil {
		p.SetCategory = true
		if string(body.CategoryID) != jsonNull {
			var id string
			if err := json.Unmarshal(body.CategoryID, &id); err != nil {
				return p, err
			}
			p.CategoryID = &id
		}
	}
	if body.StartedAt != nil {
		p.SetStarted = true
		if string(body.StartedAt) != jsonNull {
			var d string
			if err := json.Unmarshal(body.StartedAt, &d); err != nil {
				return p, err
			}
			p.StartedAt = &d
		}
	}
	if body.Attrs != nil {
		p.SetAttrs = true
		var attrs map[string]any
		if err := json.Unmarshal(body.Attrs, &attrs); err != nil {
			return p, err
		}
		p.Attrs = attrs
	}
	return p, nil
}

func parseOptionalInt(raw json.RawMessage) (*int, bool, error) {
	if raw == nil {
		return nil, false, nil
	}
	if string(raw) == jsonNull {
		return nil, true, nil
	}
	var n int
	if err := json.Unmarshal(raw, &n); err != nil {
		return nil, false, err
	}
	return &n, true, nil
}

func itemFilterFromQuery(q url.Values) (model.ItemFilter, error) {
	var costFrom, costTo *int
	if v := q.Get("cost_from"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return model.ItemFilter{}, model.Validation("invalid cost_from", map[string]any{"cost_from": detailInt})
		}
		costFrom = &n
	}
	if v := q.Get("cost_to"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return model.ItemFilter{}, model.Validation("invalid cost_to", map[string]any{"cost_to": detailInt})
		}
		costTo = &n
	}
	return model.ItemFilter{
		Q: q.Get("q"), KindID: q.Get("kind_id"), Status: q.Get("status"),
		CategoryID: q.Get("category_id"), Vendor: q.Get("vendor"),
		ExpiresFrom: q.Get("expires_from"), ExpiresTo: q.Get("expires_to"),
		CostFrom: costFrom, CostTo: costTo, BillingPeriod: q.Get("billing_period"),
		Tag: q.Get("tag"), Sort: q.Get("sort"), Order: q.Get("order"),
	}, nil
}

func parseDryRun(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "false", "0":
		return false, nil
	case "true", "1":
		return true, nil
	default:
		return false, model.Validation("invalid dry_run", map[string]any{"dry_run": "true|false"})
	}
}

func queryPage(pageRaw, perRaw string) (model.Page, error) {
	page, per := 0, 0
	var err error
	if pageRaw != "" {
		page, err = strconv.Atoi(pageRaw)
		if err != nil {
			return model.Page{}, model.Validation("invalid page", map[string]any{"page": detailInt})
		}
	}
	if perRaw != "" {
		per, err = strconv.Atoi(perRaw)
		if err != nil {
			return model.Page{}, model.Validation("invalid per_page", map[string]any{"per_page": detailInt})
		}
	}
	return service.NormalizePage(page, per)
}
