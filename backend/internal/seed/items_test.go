package seed_test

import (
	"testing"
	"time"

	"duekeep/internal/seed"
)

func TestStatusAtWrite(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)
	const (
		expired  = "expired"
		expiring = "expiring"
		active   = "active"
	)
	cases := []struct {
		name   string
		expire time.Time
		notify int
		want   string
	}{
		{"past day", today.AddDate(0, 0, -1), 30, expired},
		{"expires today", today, 30, expiring},
		{"within notify", today.AddDate(0, 0, 12), 30, expiring},
		{"on notify edge", today.AddDate(0, 0, 30), 30, expiring},
		{"after notify", today.AddDate(0, 0, 31), 30, active},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := seed.StatusAtWrite(today, tc.expire, tc.notify)
			if got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}
