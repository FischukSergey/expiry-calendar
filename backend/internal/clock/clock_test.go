package clock_test

import (
	"testing"
	"time"

	"duekeep/internal/clock"
)

func TestTodayUTC(t *testing.T) {
	t.Parallel()
	clk := clock.Fixed{T: time.Date(2026, 8, 26, 15, 4, 5, 0, time.FixedZone("MSK", 3*3600))}
	got := clock.Today(clk)
	want := clock.DateUTC(26, time.August, 2026)
	if !got.Equal(want) {
		t.Fatalf("Today = %s, want %s", got, want)
	}
	if got.Location() != time.UTC {
		t.Fatalf("Today location %s, want UTC", got.Location())
	}
}
