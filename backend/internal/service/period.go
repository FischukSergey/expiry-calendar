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

// seriesStart — нижняя граница ряда. Без started_at ряд идёт от якоря, как Sprint 9.
func seriesStart(it model.Item) (time.Time, bool, error) {
	if it.StartedAt == nil || *it.StartedAt == "" {
		return time.Time{}, false, nil
	}
	d, err := parseDate(fieldStartedAt, *it.StartedAt)
	if err != nil {
		return time.Time{}, false, err
	}
	return d, true, nil
}

func beforeSeries(it model.Item, day time.Time) (bool, error) {
	start, ok, err := seriesStart(it)
	if err != nil || !ok {
		return false, err
	}
	return day.Before(start), nil
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

// isOccurrenceDate — день из ряда записи (якорь, clamp 29–31, не раньше started_at).
func isOccurrenceDate(it model.Item, day time.Time) (bool, error) {
	expires, err := parseDate(fieldExpiresAt, it.ExpiresAt)
	if err != nil {
		return false, err
	}
	day = clock.DateUTC(day.Day(), day.Month(), day.Year())
	before, err := beforeSeries(it, day)
	if err != nil {
		return false, err
	}
	if before {
		return false, nil
	}
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
	add := func(d time.Time) error {
		if d.Before(from) || !d.Before(to) {
			return nil
		}
		skip, err := beforeSeries(it, d)
		if err != nil {
			return err
		}
		if skip {
			return nil
		}
		if openOnly && occurrencePaid(it, expires, d, paid) {
			return nil
		}
		out = append(out, d)
		return nil
	}
	switch it.BillingPeriod {
	case model.BillingMonthly:
		cur := clock.DateUTC(1, from.Month(), from.Year())
		for !cur.After(to) {
			if err := add(clampDay(cur.Year(), cur.Month(), expires.Day())); err != nil {
				return nil, err
			}
			cur = cur.AddDate(0, 1, 0)
		}
	case model.BillingYearly:
		for y := from.Year() - 1; y <= to.Year(); y++ {
			if err := add(clampDay(y, expires.Month(), expires.Day())); err != nil {
				return nil, err
			}
		}
	default:
		if err := add(expires); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// nextUnpaidOccurrence — самое раннее неоплаченное вхождение ряда, в том числе раньше from.
// Нижняя граница: started_at, иначе якорь expires_at. Месяцы шагают от 1-го числа (не AddDate от 29–31).
// one_time в прошлом тоже возвращает дату, если на неё нет платежа.
func nextUnpaidOccurrence(it model.Item, from time.Time, paid map[string]struct{}) (time.Time, bool, error) {
	expires, err := parseDate(fieldExpiresAt, it.ExpiresAt)
	if err != nil {
		return time.Time{}, false, err
	}
	lower := expires
	if start, ok, err := seriesStart(it); err != nil {
		return time.Time{}, false, err
	} else if ok {
		lower = start
	}
	switch it.BillingPeriod {
	case model.BillingMonthly:
		cur := clock.DateUTC(1, lower.Month(), lower.Year())
		end := clock.DateUTC(1, from.Month(), from.Year()).AddDate(0, 24, 0)
		for !cur.After(end) {
			d := clampDay(cur.Year(), cur.Month(), expires.Day())
			cur = cur.AddDate(0, 1, 0)
			if d.Before(lower) || occurrencePaid(it, expires, d, paid) {
				continue
			}
			return d, true, nil
		}
	case model.BillingYearly:
		endYear := from.Year() + 6
		for y := lower.Year(); y <= endYear; y++ {
			d := clampDay(y, expires.Month(), expires.Day())
			if d.Before(lower) || occurrencePaid(it, expires, d, paid) {
				continue
			}
			return d, true, nil
		}
	default:
		skip, err := beforeSeries(it, expires)
		if err != nil {
			return time.Time{}, false, err
		}
		if skip || occurrencePaid(it, expires, expires, paid) {
			return time.Time{}, false, nil
		}
		return expires, true, nil
	}
	return time.Time{}, false, nil
}

func inClosedDayWindow(day, from, toInclusive time.Time) bool {
	return !day.Before(from) && !day.After(toInclusive)
}
