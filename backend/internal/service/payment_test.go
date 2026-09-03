package service_test

import (
	"testing"
	"time"

	"duekeep/internal/clock"
	"duekeep/internal/model"
	"duekeep/internal/service"
)

func TestDashboardHidesPaidOccurrenceKeepsNextMonth(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	const owner = "owner"
	store := &overviewItems{rows: []model.Item{{
		ID: "net", OwnerID: owner, Title: "Netflix", KindID: "k1", Status: model.StatusActive,
		ExpiresAt: expiresSep15, CostAmount: 799, Currency: model.CurrencyRUB,
		BillingPeriod: model.BillingMonthly,
	}}}
	pays := newMemPayments()
	if _, created, err := pays.Insert(t.Context(), model.ItemPayment{
		ItemID: "net", OwnerID: owner, Date: expiresSep15, Amount: 799, Currency: model.CurrencyRUB,
	}); err != nil || !created {
		t.Fatalf("seed pay %v created=%v", err, created)
	}
	ov := service.NewOverview(store, clock.Fixed{T: today})
	ov.SetPayments(pays)

	got, err := ov.Dashboard(t.Context(), owner)
	if err != nil {
		t.Fatal(err)
	}
	if got.ExpirationsByMonth[1].Month != "2026-09" || got.ExpirationsByMonth[1].Count != 0 {
		t.Fatalf("sep still open: %+v", got.ExpirationsByMonth[1])
	}
	if got.ExpirationsByMonth[2].Count != 1 {
		t.Fatalf("oct should stay open: %+v", got.ExpirationsByMonth[2])
	}
	if got.Counts.Expiring7 != 0 || got.Counts.Expiring30 != 0 {
		t.Fatalf("paid day in window %+v", got.Counts)
	}
	if len(got.Soonest) != 1 || got.Soonest[0].ExpiresAt != "2026-10-15" {
		t.Fatalf("soonest %+v", got.Soonest)
	}

	sep, err := ov.Calendar(t.Context(), 2026, 9, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(sep.Days) != 1 || sep.Days[0].Items[0].OccurrenceStatus != model.OccurrencePaid {
		t.Fatalf("sep cal %+v", sep.Days)
	}
	if sep.Days[0].Items[0].CostAmount != 799 {
		t.Fatalf("sep amount %+v", sep.Days[0].Items[0])
	}
	oct, err := ov.Calendar(t.Context(), 2026, 10, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(oct.Days) != 1 || oct.Days[0].Items[0].OccurrenceStatus != model.OccurrenceOpen {
		t.Fatalf("oct cal %+v", oct.Days)
	}
}
