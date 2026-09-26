package service_test

import (
	"testing"

	"duekeep/internal/service"
)

func TestCSVSafePrefixesFormula(t *testing.T) {
	t.Parallel()
	if got := service.CSVSafeForTest("=1+1"); got != "'=1+1" {
		t.Fatalf("got %q", got)
	}
	if got := service.CSVSafeForTest("+1"); got != "'+1" {
		t.Fatalf("got %q", got)
	}
	if got := service.CSVSafeForTest("plain"); got != "plain" {
		t.Fatalf("got %q", got)
	}
}
