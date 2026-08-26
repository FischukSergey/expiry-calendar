package clock

import "time"

// Clock — инжектируемые часы (seed, статус при записи, тикер).
type Clock interface {
	Now() time.Time
}

// Real — системные часы.
type Real struct{}

// Now возвращает time.Now.
func (Real) Now() time.Time {
	return time.Now()
}

// DateUTC собирает полночь UTC из дня, месяца и года (дд.мм.гггг).
func DateUTC(day int, month time.Month, year int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// Today — календарная дата Now в UTC (как в Docker-контейнере).
func Today(c Clock) time.Time {
	n := c.Now().UTC()
	return DateUTC(n.Day(), n.Month(), n.Year())
}
