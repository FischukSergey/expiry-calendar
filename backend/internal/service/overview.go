package service

import (
	"cmp"
	"context"
	"slices"
	"time"

	"duekeep/internal/clock"
	"duekeep/internal/model"
)

const (
	fieldYear        = "year"
	fieldMonth       = "month"
	soonestLimit     = 10
	expirationMonths = 6
	monthLayout      = "2006-01"
	expiringWindow7  = 7
	expiringWindow30 = 30
	monthsPerYear    = 12
)

// OverviewStore — открытые записи владельца (не cancelled/archived). Один SELECT, без N+1.
type OverviewStore interface {
	ListOpenByOwner(ctx context.Context, ownerID string) ([]model.Item, error)
}

// Overview — дашборд и календарь. Агрегаты в памяти после одного SELECT.
type Overview struct {
	items OverviewStore
	pays  PaymentStore
	clk   clock.Clock
}

// NewOverview собирает обзор. pays может быть nil (тесты без оплат вхождения).
func NewOverview(items OverviewStore, clk clock.Clock) *Overview {
	return &Overview{items: items, clk: clk}
}

// SetPayments включает overlay дат из item_payments.
func (s *Overview) SetPayments(pays PaymentStore) {
	s.pays = pays
}

// Dashboard — GET /dashboard. Только свои открытые записи.
func (s *Overview) Dashboard(ctx context.Context, ownerID string) (model.Dashboard, error) {
	items, err := s.items.ListOpenByOwner(ctx, ownerID)
	if err != nil {
		return model.Dashboard{}, err
	}
	paidIdx, err := s.paymentsByOwner(ctx, ownerID)
	if err != nil {
		return model.Dashboard{}, err
	}
	today := clock.Today(s.clk)
	out := model.Dashboard{
		UpcomingCost:       []model.UpcomingCost{},
		ExpirationsByMonth: emptyMonths(today, expirationMonths),
		CostByKind:         []model.KindCost{},
		Soonest:            []model.DashboardItem{},
	}
	monthIdx := map[string]int{}
	monthMoney := make([]map[string]int, len(out.ExpirationsByMonth))
	for i, row := range out.ExpirationsByMonth {
		monthIdx[row.Month] = i
		monthMoney[i] = map[string]int{}
	}
	costByCur := map[string]*model.UpcomingCost{}
	kindCost := map[string]*model.KindCost{}
	brief := make([]model.DashboardItem, 0, len(items))

	winStart := clock.DateUTC(1, today.Month(), today.Year())
	winEnd := winStart.AddDate(0, expirationMonths, 0)
	for _, it := range items {
		switch it.Status {
		case model.StatusActive:
			out.Counts.Active++
		case model.StatusExpired:
			out.Counts.Expired++
		}
		paid := paidDateSet(paidIdx[it.ID])
		next, ok, err := nextUnpaidOccurrence(it, today, paid)
		if err != nil {
			return model.Dashboard{}, err
		}
		if ok {
			if inClosedDayWindow(next, today, today.AddDate(0, 0, expiringWindow7)) {
				out.Counts.Expiring7++
			}
			if inClosedDayWindow(next, today, today.AddDate(0, 0, expiringWindow30)) {
				out.Counts.Expiring30++
			}
			brief = append(brief, model.DashboardItem{
				ID: it.ID, Title: it.Title, ExpiresAt: next.Format(model.DateLayout),
				Status: it.Status, KindID: it.KindID,
			})
		}
		occs, err := occurrencesInRange(it, winStart, winEnd, paid, true)
		if err != nil {
			return model.Dashboard{}, err
		}
		for _, occ := range occs {
			if i, ok := monthIdx[occ.Format(monthLayout)]; ok {
				out.ExpirationsByMonth[i].Count++
				monthMoney[i][it.Currency] += it.CostAmount
			}
		}
		if it.Status == model.StatusExpired {
			continue
		}
		addUpcoming(costByCur, it)
		key := it.KindID + "\x00" + it.Currency
		if row, ok := kindCost[key]; ok {
			row.Amount += it.CostAmount
			continue
		}
		kindCost[key] = &model.KindCost{KindID: it.KindID, Currency: it.Currency, Amount: it.CostAmount}
	}

	for i := range out.ExpirationsByMonth {
		out.ExpirationsByMonth[i].Amounts = sortedCurrencyAmounts(monthMoney[i])
	}
	out.UpcomingCost = sortedUpcoming(costByCur)
	out.CostByKind = sortedKindCost(kindCost)
	slices.SortFunc(brief, func(a, b model.DashboardItem) int {
		if n := cmp.Compare(a.ExpiresAt, b.ExpiresAt); n != 0 {
			return n
		}
		return cmp.Compare(a.ID, b.ID)
	})
	out.Soonest = brief[:min(len(brief), soonestLimit)]
	return out, nil
}

// Calendar — GET /calendar. year 1..9999, month 1..12. Только свои; пустые дни опускаем.
func (s *Overview) Calendar(ctx context.Context, year, month int, ownerID string) (model.Calendar, error) {
	if year < 1 || year > 9999 {
		return model.Calendar{}, model.Validation("invalid year", map[string]any{fieldYear: "1..9999"})
	}
	if month < 1 || month > 12 {
		return model.Calendar{}, model.Validation("invalid month", map[string]any{fieldMonth: "1..12"})
	}
	items, err := s.items.ListOpenByOwner(ctx, ownerID)
	if err != nil {
		return model.Calendar{}, err
	}
	paidIdx, err := s.paymentsByOwner(ctx, ownerID)
	if err != nil {
		return model.Calendar{}, err
	}
	start := clock.DateUTC(1, time.Month(month), year)
	end := start.AddDate(0, 1, 0)
	byDay := map[string][]model.CalendarItem{}
	for _, it := range items {
		expires, err := parseDate(fieldExpiresAt, it.ExpiresAt)
		if err != nil {
			return model.Calendar{}, err
		}
		paid := paidDateSet(paidIdx[it.ID])
		occs, err := occurrencesInRange(it, start, end, paid, false)
		if err != nil {
			return model.Calendar{}, err
		}
		for _, occ := range occs {
			day := occ.Format(model.DateLayout)
			occStatus := model.OccurrenceOpen
			amount, currency := it.CostAmount, it.Currency
			if occurrencePaid(it, expires, occ, paid) {
				occStatus = model.OccurrencePaid
				if snap, ok := paidIdx[it.ID][day]; ok {
					amount, currency = snap.Amount, snap.Currency
				}
			}
			byDay[day] = append(byDay[day], model.CalendarItem{
				ID: it.ID, Title: it.Title, Status: it.Status,
				OccurrenceStatus: occStatus, CostAmount: amount, Currency: currency,
			})
		}
	}
	days := make([]model.CalendarDay, 0, len(byDay))
	for date, list := range byDay {
		slices.SortFunc(list, func(a, b model.CalendarItem) int {
			if n := cmp.Compare(a.Title, b.Title); n != 0 {
				return n
			}
			return cmp.Compare(a.ID, b.ID)
		})
		days = append(days, model.CalendarDay{Date: date, Items: list})
	}
	slices.SortFunc(days, func(a, b model.CalendarDay) int {
		return cmp.Compare(a.Date, b.Date)
	})
	return model.Calendar{Year: year, Month: month, Days: days}, nil
}

func (s *Overview) paymentsByOwner(ctx context.Context, ownerID string) (map[string]map[string]model.ItemPayment, error) {
	if s.pays == nil {
		return map[string]map[string]model.ItemPayment{}, nil
	}
	rows, err := s.pays.ListByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return paidDatesByItem(rows), nil
}

func emptyMonths(today time.Time, n int) []model.MonthCount {
	out := make([]model.MonthCount, n)
	cur := clock.DateUTC(1, today.Month(), today.Year())
	for i := range n {
		out[i] = model.MonthCount{Month: cur.Format(monthLayout), Amounts: []model.CurrencyAmount{}}
		cur = cur.AddDate(0, 1, 0)
	}
	return out
}

func sortedCurrencyAmounts(byCur map[string]int) []model.CurrencyAmount {
	out := make([]model.CurrencyAmount, 0, len(byCur))
	for cur, amt := range byCur {
		if amt == 0 {
			continue
		}
		out = append(out, model.CurrencyAmount{Currency: cur, Amount: amt})
	}
	slices.SortFunc(out, func(a, b model.CurrencyAmount) int {
		return cmp.Compare(a.Currency, b.Currency)
	})
	return out
}

func addUpcoming(byCur map[string]*model.UpcomingCost, it model.Item) {
	row, ok := byCur[it.Currency]
	if !ok {
		row = &model.UpcomingCost{Currency: it.Currency}
		byCur[it.Currency] = row
	}
	switch it.BillingPeriod {
	case model.BillingMonthly:
		row.Monthly += it.CostAmount
		row.Yearly += it.CostAmount * monthsPerYear
	case model.BillingYearly:
		row.Yearly += it.CostAmount
		row.Monthly += it.CostAmount / monthsPerYear
	}
}

func sortedUpcoming(byCur map[string]*model.UpcomingCost) []model.UpcomingCost {
	out := make([]model.UpcomingCost, 0, len(byCur))
	for _, row := range byCur {
		out = append(out, *row)
	}
	slices.SortFunc(out, func(a, b model.UpcomingCost) int {
		return cmp.Compare(a.Currency, b.Currency)
	})
	return out
}

func sortedKindCost(byKey map[string]*model.KindCost) []model.KindCost {
	out := make([]model.KindCost, 0, len(byKey))
	for _, row := range byKey {
		out = append(out, *row)
	}
	slices.SortFunc(out, func(a, b model.KindCost) int {
		if n := cmp.Compare(a.KindID, b.KindID); n != 0 {
			return n
		}
		return cmp.Compare(a.Currency, b.Currency)
	})
	return out
}
