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

func TestDefaultCategoriesParentOrder(t *testing.T) {
	t.Parallel()
	cats := seed.DefaultCategories()
	if len(cats) < 10 {
		t.Fatalf("want at least 10, got %d", len(cats))
	}
	names := make(map[string]struct{}, len(cats))
	for i, c := range cats {
		if c.Name == "" {
			t.Fatalf("empty name at %d", i)
		}
		if _, dup := names[c.Name]; dup {
			t.Fatalf("duplicate name %s", c.Name)
		}
		names[c.Name] = struct{}{}
		if c.ParentIdx < -1 || c.ParentIdx >= i {
			t.Fatalf("%s: ParentIdx %d", c.Name, c.ParentIdx)
		}
	}
}
