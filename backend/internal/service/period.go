package service

import (
	"time"

	"duekeep/internal/clock"
	"duekeep/internal/model"
)

// clampDay ставит день месяца; 29–31 укорачивает до последнего дня месяца.
func clampDay(year int, month time.Month, day int) time.Time {
	last := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	return clock.DateUTC(min(day, last), month, year)
}

func skipPaidOccurrence(it model.Item, expires, day time.Time) bool {
	return it.Status == model.StatusPaid && !day.After(expires)
}

func occurrencePaid(it model.Item, expires, day time.Time, paid map[string]struct{}) bool {
	if skipPaidOccurrence(it, expires, day) {
		return true
	}
	if paid == nil {
		return false
	}
	_, ok := paid[day.Format(model.DateLayout)]
	return ok
}

// isOccurrenceDate — день из ряда записи (якорь и clamp 29–31, как развёртка).
func isOccurrenceDate(it model.Item, day time.Time) (bool, error) {
	expires, err := parseDate(fieldExpiresAt, it.ExpiresAt)
	if err != nil {
		return false, err
	}
	day = clock.DateUTC(day.Day(), day.Month(), day.Year())
	switch it.BillingPeriod {
	case model.BillingMonthly:
		return sameDay(clampDay(day.Year(), day.Month(), expires.Day()), day), nil
	case model.BillingYearly:
		return sameDay(clampDay(day.Year(), expires.Month(), expires.Day()), day), nil
	default:
		return sameDay(expires, day), nil
	}
}

func sameDay(a, b time.Time) bool {
	return a.Equal(b)
}

// occurrencesInRange — вхождения периода в [from, to).
// openOnly: без заморозки paid и без дат из item_payments (обзор / «сгорит»).
func occurrencesInRange(it model.Item, from, to time.Time, paid map[string]struct{}, openOnly bool) ([]time.Time, error) {
	expires, err := parseDate(fieldExpiresAt, it.ExpiresAt)
	if err != nil {
		return nil, err
	}
	var out []time.Time
	add := func(d time.Time) {
		if d.Before(from) || !d.Before(to) {
			return
		}
		if openOnly && occurrencePaid(it, expires, d, paid) {
			return
		}
		out = append(out, d)
	}
	switch it.BillingPeriod {
	case model.BillingMonthly:
		cur := clock.DateUTC(1, from.Month(), from.Year())
		for !cur.After(to) {
			add(clampDay(cur.Year(), cur.Month(), expires.Day()))
			cur = cur.AddDate(0, 1, 0)
		}
	case model.BillingYearly:
		for y := from.Year() - 1; y <= to.Year(); y++ {
			add(clampDay(y, expires.Month(), expires.Day()))
		}
	default:
		add(expires)
	}
	return out, nil
}

// nextUnpaidOccurrence — ближайшее вхождение ≥ from, не скрытое paid и не в item_payments.
// one_time в прошлом тоже возвращает дату, если на неё нет платежа.
func nextUnpaidOccurrence(it model.Item, from time.Time, paid map[string]struct{}) (time.Time, bool, error) {
	expires, err := parseDate(fieldExpiresAt, it.ExpiresAt)
	if err != nil {
		return time.Time{}, false, err
	}
	switch it.BillingPeriod {
	case model.BillingMonthly:
		for i := range 24 {
			probe := from.AddDate(0, i, 0)
			d := clampDay(probe.Year(), probe.Month(), expires.Day())
			if d.Before(from) || occurrencePaid(it, expires, d, paid) {
				continue
			}
			return d, true, nil
		}
	case model.BillingYearly:
		for i := range 6 {
			d := clampDay(from.Year()+i, expires.Month(), expires.Day())
			if d.Before(from) || occurrencePaid(it, expires, d, paid) {
				continue
			}
			return d, true, nil
		}
	default:
		if occurrencePaid(it, expires, expires, paid) {
			return time.Time{}, false, nil
		}
		return expires, true, nil
	}
	return time.Time{}, false, nil
}

func inClosedDayWindow(day, from, toInclusive time.Time) bool {
	return !day.Before(from) && !day.After(toInclusive)
}
