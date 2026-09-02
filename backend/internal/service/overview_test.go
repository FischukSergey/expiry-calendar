package service_test

import (
	"context"
	"testing"
	"time"

	"duekeep/internal/clock"
	"duekeep/internal/model"
	"duekeep/internal/service"
)

type overviewItems struct {
	rows []model.Item
}

func (m *overviewItems) ListOpenByOwner(_ context.Context, ownerID string) ([]model.Item, error) {
	out := make([]model.Item, 0, len(m.rows))
	for _, it := range m.rows {
		if it.Status == model.StatusCancelled || it.Status == model.StatusArchived {
			continue
		}
		if it.OwnerID != ownerID {
			continue
		}
		out = append(out, it)
	}
	return out, nil
}

func TestDashboardTwoCurrenciesAndSkipsCancelled(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	const owner = "owner"
	store := &overviewItems{rows: []model.Item{
		{
			ID: "a", OwnerID: owner, Title: itemTitleDomain, KindID: "k1", Status: model.StatusActive,
			ExpiresAt: expiresSep, CostAmount: 100, Currency: model.CurrencyRUB,
			BillingPeriod: model.BillingMonthly,
		},
		{
			ID: "b", OwnerID: owner, Title: "SaaS", KindID: "k2", Status: model.StatusExpiring,
			ExpiresAt: "2026-08-28", CostAmount: 120, Currency: "USD",
			BillingPeriod: model.BillingYearly,
		},
		{
			ID: "c", OwnerID: owner, Title: "Отмена", KindID: "k1", Status: model.StatusCancelled,
			ExpiresAt: "2026-08-27", CostAmount: 999, Currency: model.CurrencyRUB,
			BillingPeriod: model.BillingMonthly,
		},
		{
			ID: "d", OwnerID: owner, Title: "Просрочка", KindID: "k1", Status: model.StatusExpired,
			ExpiresAt: expiresPast, CostAmount: 50, Currency: model.CurrencyRUB,
			BillingPeriod: model.BillingMonthly,
		},
		{
			ID: "x", OwnerID: otherOwner, Title: "Чужое", KindID: "k1", Status: model.StatusActive,
			ExpiresAt: expiresSep, CostAmount: 5000, Currency: model.CurrencyRUB,
			BillingPeriod: model.BillingMonthly,
		},
	}}
	ov := service.NewOverview(store, clock.Fixed{T: today})
	got, err := ov.Dashboard(t.Context(), owner)
	if err != nil {
		t.Fatal(err)
	}
	if got.Counts.Active != 1 || got.Counts.Expired != 1 {
		t.Fatalf("counts %+v", got.Counts)
	}
	if got.Counts.Expiring7 != 3 || got.Counts.Expiring30 != 3 {
		t.Fatalf("windows %+v", got.Counts)
	}
	if len(got.UpcomingCost) != 2 {
		t.Fatalf("currencies %d %+v", len(got.UpcomingCost), got.UpcomingCost)
	}
	rub := got.UpcomingCost[0]
	if rub.Currency != model.CurrencyRUB || rub.Monthly != 100 || rub.Yearly != 1200 {
		t.Fatalf("RUB %+v", rub)
	}
	if got.UpcomingCost[1].Currency != "USD" || got.UpcomingCost[1].Yearly != 120 || got.UpcomingCost[1].Monthly != 10 {
		t.Fatalf("USD %+v", got.UpcomingCost[1])
	}
	if len(got.CostByKind) != 2 {
		t.Fatalf("kinds %+v", got.CostByKind)
	}
	if got.Soonest[0].ID != "b" || len(got.Soonest) != 3 {
		t.Fatalf("soonest %+v", got.Soonest)
	}
	if len(got.ExpirationsByMonth) != 6 || got.ExpirationsByMonth[0].Month != "2026-08" {
		t.Fatalf("months %+v", got.ExpirationsByMonth)
	}
	if got.ExpirationsByMonth[0].Count != 3 || got.ExpirationsByMonth[1].Count != 2 {
		t.Fatalf("month counts %+v", got.ExpirationsByMonth)
	}
	if monthAmount(got.ExpirationsByMonth[0], model.CurrencyRUB) != 150 ||
		monthAmount(got.ExpirationsByMonth[0], "USD") != 120 {
		t.Fatalf("aug amounts %+v", got.ExpirationsByMonth[0].Amounts)
	}
	if monthAmount(got.ExpirationsByMonth[1], model.CurrencyRUB) != 150 {
		t.Fatalf("sep amounts %+v", got.ExpirationsByMonth[1].Amounts)
	}
}

func monthAmount(row model.MonthCount, currency string) int {
	for _, a := range row.Amounts {
		if a.Currency == currency {
			return a.Amount
		}
	}
	return 0
}

func TestCalendarMonthAndEmptyDays(t *testing.T) {
	t.Parallel()
	const owner = "owner"
	store := &overviewItems{rows: []model.Item{
		{ID: "a", OwnerID: owner, Title: "Бета", Status: model.StatusExpiring, ExpiresAt: calendarDay},
		{ID: "b", OwnerID: owner, Title: "Альфа", Status: model.StatusActive, ExpiresAt: calendarDay},
		{ID: "c", OwnerID: owner, Title: "Сентябрь", Status: model.StatusActive, ExpiresAt: expiresSep},
		{ID: "x", OwnerID: otherOwner, Title: "Чужое", Status: model.StatusActive, ExpiresAt: calendarDay},
	}}
	ov := service.NewOverview(store, clock.Fixed{T: time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)})
	got, err := ov.Calendar(t.Context(), 2026, 8, owner)
	if err != nil {
		t.Fatal(err)
	}
	if got.Year != 2026 || got.Month != 8 || len(got.Days) != 1 {
		t.Fatalf("cal %+v", got)
	}
	if got.Days[0].Date != calendarDay || len(got.Days[0].Items) != 2 {
		t.Fatalf("day %+v", got.Days[0])
	}
	if got.Days[0].Items[0].Title != "Альфа" {
		t.Fatalf("sort %+v", got.Days[0].Items)
	}
	if _, err := ov.Calendar(t.Context(), 2026, 13, owner); err == nil {
		t.Fatal("month 13")
	}
}

func TestDashboardMonthlyExpansionAndPaid(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	const owner = "owner"
	store := &overviewItems{rows: []model.Item{
		{
			ID: "far", OwnerID: owner, Title: "Netflix", KindID: "k1", Status: model.StatusActive,
			ExpiresAt: "2027-03-15", CostAmount: 799, Currency: model.CurrencyRUB,
			BillingPeriod: model.BillingMonthly,
		},
		{
			ID: "paid", OwnerID: owner, Title: "Офис", KindID: "k1", Status: model.StatusPaid,
			ExpiresAt: "2026-08-28", CostAmount: 10000, Currency: model.CurrencyRUB,
			BillingPeriod: model.BillingMonthly,
		},
		{
			ID: "once", OwnerID: owner, Title: "Разово", KindID: "k2", Status: model.StatusPaid,
			ExpiresAt: "2026-08-27", CostAmount: 50, Currency: model.CurrencyRUB,
			BillingPeriod: model.BillingOneTime,
		},
	}}
	ov := service.NewOverview(store, clock.Fixed{T: today})
	got, err := ov.Dashboard(t.Context(), owner)
	if err != nil {
		t.Fatal(err)
	}
	if got.ExpirationsByMonth[0].Count != 1 {
		t.Fatalf("aug should hide paid current: %+v", got.ExpirationsByMonth[0])
	}
	if monthAmount(got.ExpirationsByMonth[0], model.CurrencyRUB) != 799 {
		t.Fatalf("aug amount %+v", got.ExpirationsByMonth[0].Amounts)
	}
	if got.ExpirationsByMonth[1].Count != 2 {
		t.Fatalf("sep wants far+next paid: %+v", got.ExpirationsByMonth[1])
	}
	if got.Counts.Expiring7 != 0 {
		t.Fatalf("paid current not in window: %+v", got.Counts)
	}
	foundFar := false
	for _, row := range got.Soonest {
		if row.ID == "once" {
			t.Fatal("paid one_time in soonest")
		}
		if row.ID == "paid" && row.ExpiresAt != "2026-09-28" {
			t.Fatalf("paid next %+v", row)
		}
		if row.ID == "far" {
			foundFar = true
			if row.ExpiresAt != "2026-09-15" {
				t.Fatalf("far date %s", row.ExpiresAt)
			}
		}
	}
	if !foundFar {
		t.Fatal("far monthly missing soonest")
	}

	cal, err := ov.Calendar(t.Context(), 2026, 9, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(cal.Days) != 2 {
		t.Fatalf("sep days %+v", cal.Days)
	}
	aug, err := ov.Calendar(t.Context(), 2026, 8, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(aug.Days) != 1 || aug.Days[0].Date != "2026-08-15" {
		t.Fatalf("aug cal %+v", aug.Days)
	}

	clampStore := &overviewItems{rows: []model.Item{
		{
			ID: "jan31", OwnerID: owner, Title: "Clamp", Status: model.StatusActive,
			ExpiresAt: "2026-01-31", BillingPeriod: model.BillingMonthly,
		},
	}}
	ovClamp := service.NewOverview(clampStore, clock.Fixed{T: today})
	feb, err := ovClamp.Calendar(t.Context(), 2026, 2, owner)
	if err != nil {
		t.Fatal(err)
	}
	if len(feb.Days) != 1 || feb.Days[0].Date != "2026-02-28" {
		t.Fatalf("clamp feb %+v", feb.Days)
	}
}
