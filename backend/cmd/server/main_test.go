package main

import "testing"

func TestSeedEnabled(t *testing.T) {
	t.Parallel()
	cases := []struct {
		raw  string
		want bool
	}{
		{"", false},
		{envTrue, true},
		{"1", true},
		{"yes", true},
		{"on", true},
		{"false", false},
		{"FALSE", false},
		{"0", false},
		{"no", false},
		{"off", false},
		{" off ", false},
	}
	for _, tc := range cases {
		if got := seedEnabled(tc.raw); got != tc.want {
			t.Fatalf("seedEnabled(%q)=%v want %v", tc.raw, got, tc.want)
		}
	}
}
