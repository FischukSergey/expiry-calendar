package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"duekeep/internal/model"
)

func TestPayUnpayIdempotentAndErrors(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	tok := testJWT(t, string(model.RoleAdmin))
	it := adminCreateItem(t, api, tok, `{"title":"`+itemTitleDomain+`","kind_id":"`+otherKindID+
		`","expires_at":"2026-09-15","billing_period":"monthly","cost_amount":799}`)

	rec := adminJSON(t, api, tok, http.MethodPost, "/api/v1/items/"+it.ID+"/payments",
		`{"date":"2026-09-15"}`, http.StatusCreated)
	var pay model.ItemPayment
	if err := json.NewDecoder(rec.Body).Decode(&pay); err != nil {
		t.Fatal(err)
	}
	if pay.Date != "2026-09-15" || pay.Amount != 799 || pay.Currency != model.CurrencyRUB || pay.ItemID != it.ID {
		t.Fatalf("pay %+v", pay)
	}

	again := adminJSON(t, api, tok, http.MethodPost, "/api/v1/items/"+it.ID+"/payments",
		`{"date":"2026-09-15"}`, http.StatusOK)
	var same model.ItemPayment
	if err := json.NewDecoder(again.Body).Decode(&same); err != nil {
		t.Fatal(err)
	}
	if same.ID != pay.ID {
		t.Fatalf("idempotent %s %s", same.ID, pay.ID)
	}

	adminJSON(t, api, tok, http.MethodPost, "/api/v1/items/"+it.ID+"/payments",
		`{"date":"2026-09-16"}`, http.StatusUnprocessableEntity)

	other := testJWTSub(t, string(model.RoleAdmin), "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	adminJSON(t, api, other, http.MethodPost, "/api/v1/items/"+it.ID+"/payments",
		`{"date":"2026-10-15"}`, http.StatusNotFound)

	cardRec := adminJSON(t, api, tok, http.MethodGet, "/api/v1/items/"+it.ID, "", http.StatusOK)
	var card model.ItemCard
	if err := json.NewDecoder(cardRec.Body).Decode(&card); err != nil {
		t.Fatal(err)
	}
	if card.NextOpenAt == nil || *card.NextOpenAt != "2026-10-15" {
		t.Fatalf("next_open_at %+v", card.NextOpenAt)
	}

	adminJSON(t, api, tok, http.MethodDelete, "/api/v1/items/"+it.ID+"/payments?date=2026-09-15", "", http.StatusNoContent)
	adminJSON(t, api, tok, http.MethodDelete, "/api/v1/items/"+it.ID+"/payments?date=2026-09-15", "", http.StatusNoContent)
	adminJSON(t, api, tok, http.MethodDelete, "/api/v1/items/"+it.ID+"/payments", "", http.StatusUnprocessableEntity)

	auditRec := adminJSON(t, api, tok, http.MethodGet, "/api/v1/audit?per_page=20", "", http.StatusOK)
	var audit model.AuditList
	if err := json.NewDecoder(auditRec.Body).Decode(&audit); err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, e := range audit.Items {
		got[e.Action]++
		assertAuditNoSecrets(t, e)
	}
	if got[model.AuditPay] < 1 || got[model.AuditUnpay] < 1 {
		t.Fatalf("audit pay/unpay %+v", got)
	}
}

func TestPayClearsExpiredAndRejectsBeforeStart(t *testing.T) {
	t.Parallel()
	api := itemsAPI(t)
	tok := testJWT(t, string(model.RoleAdmin))
	expired := adminCreateItem(t, api, tok, `{"title":"`+itemTitleDomain+`","kind_id":"`+otherKindID+
		`","expires_at":"2026-08-01","billing_period":"one_time"}`)
	if expired.Status != model.StatusExpired {
		t.Fatalf("create status %s", expired.Status)
	}
	adminJSON(t, api, tok, http.MethodPost, "/api/v1/items/"+expired.ID+"/payments",
		`{"date":"2026-08-01"}`, http.StatusCreated)
	listRec := adminJSON(t, api, tok, http.MethodGet, "/api/v1/items/"+expired.ID, "", http.StatusOK)
	var card model.ItemCard
	if err := json.NewDecoder(listRec.Body).Decode(&card); err != nil {
		t.Fatal(err)
	}
	if card.Item.Status != model.StatusActive {
		t.Fatalf("after pay status %s", card.Item.Status)
	}
	listed := listItems(t, api, tok, "q="+itemTitleDomain)
	found := false
	for _, it := range listed.Items {
		if it.ID == expired.ID {
			found = true
			if it.Status != model.StatusActive {
				t.Fatalf("list status %s", it.Status)
			}
		}
	}
	if !found {
		t.Fatal("paid item missing from list")
	}

	monthly := adminCreateItem(t, api, tok, `{"title":"Подписка","kind_id":"`+otherKindID+
		`","expires_at":"2026-08-01","billing_period":"monthly"}`)
	if monthly.Status == model.StatusExpired {
		t.Fatalf("monthly past anchor %s", monthly.Status)
	}

	start := adminCreateItem(t, api, tok, `{"title":"Аренда","kind_id":"`+otherKindID+
		`","expires_at":"2026-10-15","started_at":"2026-10-01","billing_period":"monthly"}`)
	if start.Status != model.StatusActive {
		t.Fatalf("future start status %s", start.Status)
	}
	adminJSON(t, api, tok, http.MethodPost, "/api/v1/items/"+start.ID+"/payments",
		`{"date":"2026-08-15"}`, http.StatusUnprocessableEntity)
	adminJSON(t, api, tok, http.MethodPost, "/api/v1/items/"+start.ID+"/payments",
		`{"date":"2026-10-15"}`, http.StatusCreated)
}
