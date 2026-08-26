package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"duekeep/internal/middleware"
	"duekeep/internal/model"
	"duekeep/internal/service"
)

type itemWrite struct {
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	KindID           string         `json:"kind_id"`
	CategoryID       *string        `json:"category_id"`
	Vendor           string         `json:"vendor"`
	Tags             []string       `json:"tags"`
	CostAmount       *int           `json:"cost_amount"`
	Currency         string         `json:"currency"`
	BillingPeriod    string         `json:"billing_period"`
	StartedAt        *string        `json:"started_at"`
	ExpiresAt        string         `json:"expires_at"`
	NotifyBeforeDays *int           `json:"notify_before_days"`
	URL              string         `json:"url"`
	AccountHint      string         `json:"account_hint"`
	Status           string         `json:"status"`
	Attrs            map[string]any `json:"attrs"`
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
	NotifyBeforeDays *int            `json:"notify_before_days"`
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
	var costFrom, costTo *int
	if v := q.Get("cost_from"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid cost_from")
			return
		}
		costFrom = &n
	}
	if v := q.Get("cost_to"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid cost_to")
			return
		}
		costTo = &n
	}
	out, err := a.items.List(r.Context(), model.ItemFilter{
		Q: q.Get("q"), KindID: q.Get("kind_id"), Status: q.Get("status"),
		CategoryID: q.Get("category_id"), Vendor: q.Get("vendor"),
		ExpiresFrom: q.Get("expires_from"), ExpiresTo: q.Get("expires_to"),
		CostFrom: costFrom, CostTo: costTo, BillingPeriod: q.Get("billing_period"),
		Tag: q.Get("tag"), Sort: q.Get("sort"), Order: q.Get("order"),
	}, page)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, out)
}

func (a *API) createItem(w http.ResponseWriter, r *http.Request) {
	var body itemWrite
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "invalid json")
		return
	}
	it, err := a.items.Create(r.Context(), itemFromWrite(body), middleware.UserID(r.Context()))
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
	card, err := a.items.Get(r.Context(), id)
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
	out, err := a.items.ListAudit(r.Context(), page)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeBytes(w, http.StatusOK, out)
}

func itemFromWrite(body itemWrite) model.Item {
	cost := 0
	if body.CostAmount != nil {
		cost = *body.CostAmount
	}
	notify := model.DefaultNotifyDays
	if body.NotifyBeforeDays != nil {
		notify = *body.NotifyBeforeDays
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
	}
}

func itemPatchFromBody(body itemPatchBody) (model.ItemPatch, error) {
	p := model.ItemPatch{
		Title: body.Title, Description: body.Description, KindID: body.KindID,
		Vendor: body.Vendor, Tags: body.Tags, CostAmount: body.CostAmount,
		Currency: body.Currency, BillingPeriod: body.BillingPeriod, ExpiresAt: body.ExpiresAt,
		NotifyBeforeDays: body.NotifyBeforeDays, URL: body.URL, AccountHint: body.AccountHint,
		Status: body.Status,
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

func queryPage(pageRaw, perRaw string) (model.Page, error) {
	page, per := 0, 0
	var err error
	if pageRaw != "" {
		page, err = strconv.Atoi(pageRaw)
		if err != nil {
			return model.Page{}, model.Validation("invalid page", map[string]any{"page": "int"})
		}
	}
	if perRaw != "" {
		per, err = strconv.Atoi(perRaw)
		if err != nil {
			return model.Page{}, model.Validation("invalid per_page", map[string]any{"per_page": "int"})
		}
	}
	return service.NormalizePage(page, per)
}
