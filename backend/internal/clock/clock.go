package clock

import "time"

// Clock — инжектируемые часы (seed и тикер в следующих спринтах).
type Clock interface {
	Now() time.Time
}

// Real — системные часы.
type Real struct{}

// Now возвращает time.Now.
func (Real) Now() time.Time {
	return time.Now()
}
