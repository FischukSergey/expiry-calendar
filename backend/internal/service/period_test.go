package service_test

import (
	"testing"
	"time"

	"duekeep/internal/model"
	"duekeep/internal/service"
)

const anchorJan15 = "2026-01-15"
const anchorJan31 = "2026-01-31"

func TestNextUnpaidKeepsFebruaryWhenTodayIs31(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	it := model.Item{
		ExpiresAt:        anchorJan31,
		BillingPeriod:    model.BillingMonthly,
		NotifyBeforeDays: model.Ptr(30),
	}
	paid := map[string]struct{}{anchorJan31: {}}
	got, ok, err := service.NextUnpaidForTest(it, today, paid)
	if err != nil || !ok {
		t.Fatalf("next %v %v", got, err)
	}
	if got.Format(model.DateLayout) != "2026-02-28" {
		t.Fatalf("february skipped: %s", got.Format(model.DateLayout))
	}
}

func TestMonthlyPastUnpaidStaysExpiredUntilPaid(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	it := model.Item{
		ExpiresAt:        anchorJan15,
		BillingPeriod:    model.BillingMonthly,
		NotifyBeforeDays: model.Ptr(30),
		Status:           model.StatusActive,
	}
	status, err := service.StatusFromOccurrencesForTest(it, today, nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != model.StatusExpired {
		t.Fatalf("status %s", status)
	}
	next, ok, err := service.NextUnpaidForTest(it, today, nil)
	if err != nil || !ok || next.Format(model.DateLayout) != anchorJan15 {
		t.Fatalf("next %s ok=%v err=%v", next.Format(model.DateLayout), ok, err)
	}

	paid := map[string]struct{}{anchorJan15: {}}
	status, err = service.StatusFromOccurrencesForTest(it, today, paid)
	if err != nil {
		t.Fatal(err)
	}
	if status != model.StatusExpired {
		t.Fatalf("february still open: %s", status)
	}
	paid["2026-02-15"] = struct{}{}
	paid["2026-03-15"] = struct{}{}
	status, err = service.StatusFromOccurrencesForTest(it, today, paid)
	if err != nil {
		t.Fatal(err)
	}
	if status != model.StatusExpiring {
		t.Fatalf("april should be upcoming: %s", status)
	}
	next, ok, err = service.NextUnpaidForTest(it, today, paid)
	if err != nil || !ok || next.Format(model.DateLayout) != "2026-04-15" {
		t.Fatalf("next %s ok=%v", next.Format(model.DateLayout), ok)
	}
}

func TestPaidFreezeIgnoresPastOccurrence(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	it := model.Item{
		ExpiresAt:     anchorJan15,
		BillingPeriod: model.BillingMonthly,
		Status:        model.StatusPaid,
	}
	status, err := service.StatusFromOccurrencesForTest(it, today, nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != model.StatusPaid {
		t.Fatalf("status %s", status)
	}
}
