package clock

import "time"

// Fixed отдаёт одну и ту же метку — для тестов JWT и TTL.
type Fixed struct {
	T time.Time
}

// Now возвращает зафиксированное T.
func (f Fixed) Now() time.Time {
	return f.T
}
