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

func (m *overviewItems) ListOpen(context.Context) ([]model.Item, error) {
	out := make([]model.Item, 0, len(m.rows))
	for _, it := range m.rows {
		if it.Status == model.StatusCancelled || it.Status == model.StatusArchived {
			continue
		}
		out = append(out, it)
	}
	return out, nil
}

func TestDashboardTwoCurrenciesAndSkipsCancelled(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
	store := &overviewItems{rows: []model.Item{
		{
			ID: "a", Title: itemTitleDomain, KindID: "k1", Status: model.StatusActive,
			ExpiresAt: "2026-09-01", CostAmount: 100, Currency: model.CurrencyRUB,
			BillingPeriod: model.BillingMonthly,
		},
		{
			ID: "b", Title: "SaaS", KindID: "k2", Status: model.StatusExpiring,
			ExpiresAt: "2026-08-28", CostAmount: 120, Currency: "USD",
			BillingPeriod: model.BillingYearly,
		},
		{
			ID: "c", Title: "Отмена", KindID: "k1", Status: model.StatusCancelled,
			ExpiresAt: "2026-08-27", CostAmount: 999, Currency: model.CurrencyRUB,
			BillingPeriod: model.BillingMonthly,
		},
		{
			ID: "d", Title: "Просрочка", KindID: "k1", Status: model.StatusExpired,
			ExpiresAt: expiresPast, CostAmount: 50, Currency: model.CurrencyRUB,
			BillingPeriod: model.BillingMonthly,
		},
	}}
	ov := service.NewOverview(store, clock.Fixed{T: today})
	got, err := ov.Dashboard(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if got.Counts.Active != 1 || got.Counts.Expired != 1 {
		t.Fatalf("counts %+v", got.Counts)
	}
	if got.Counts.Expiring7 != 2 || got.Counts.Expiring30 != 2 {
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
	if got.Soonest[0].ID != "d" || len(got.Soonest) != 3 {
		t.Fatalf("soonest %+v", got.Soonest)
	}
	if len(got.ExpirationsByMonth) != 6 || got.ExpirationsByMonth[0].Month != "2026-08" {
		t.Fatalf("months %+v", got.ExpirationsByMonth)
	}
	if got.ExpirationsByMonth[0].Count != 2 || got.ExpirationsByMonth[1].Count != 1 {
		t.Fatalf("month counts %+v", got.ExpirationsByMonth)
	}
}

func TestCalendarMonthAndEmptyDays(t *testing.T) {
	t.Parallel()
	store := &overviewItems{rows: []model.Item{
		{ID: "a", Title: "Бета", Status: model.StatusExpiring, ExpiresAt: calendarDay},
		{ID: "b", Title: "Альфа", Status: model.StatusActive, ExpiresAt: calendarDay},
		{ID: "c", Title: "Сентябрь", Status: model.StatusActive, ExpiresAt: "2026-09-01"},
	}}
	ov := service.NewOverview(store, clock.Fixed{T: time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)})
	got, err := ov.Calendar(t.Context(), 2026, 8)
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
	if _, err := ov.Calendar(t.Context(), 2026, 13); err == nil {
		t.Fatal("month 13")
	}
}
