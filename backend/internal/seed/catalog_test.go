package seed_test

import (
	"testing"

	"duekeep/internal/seed"
)

func TestCheckCatalog(t *testing.T) {
	t.Parallel()
	if err := seed.CheckCatalog(); err != nil {
		t.Fatal(err)
	}
}
