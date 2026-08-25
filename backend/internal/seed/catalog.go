package seed

import (
	"fmt"
	"slices"
)

var requiredKindSlugs = []string{
	"domain", "subscription", "rent", "contract", "insurance",
	"license", "tax", "vehicle", "other",
}

var forbiddenKindSlugs = []string{"ssl", "warranty"}

// CheckCatalog проверяет инварианты seed-справочников без обращения к БД.
func CheckCatalog() error {
	if len(userSeeds) != 2 {
		return fmt.Errorf("want 2 users, got %d", len(userSeeds))
	}
	if len(kindSeeds) != len(requiredKindSlugs) {
		return fmt.Errorf("want %d kinds, got %d", len(requiredKindSlugs), len(kindSeeds))
	}

	slugs := make([]string, 0, len(kindSeeds))
	kindIDs := make(map[string]struct{}, len(kindSeeds))
	for _, k := range kindSeeds {
		if _, dup := kindIDs[k.id]; dup {
			return fmt.Errorf("duplicate kind id %s", k.id)
		}
		kindIDs[k.id] = struct{}{}
		slugs = append(slugs, k.slug)
		if err := checkAttrSchema(k.slug, k.attrSchema); err != nil {
			return err
		}
	}
	for _, slug := range requiredKindSlugs {
		if !slices.Contains(slugs, slug) {
			return fmt.Errorf("missing kind %s", slug)
		}
	}
	for _, slug := range forbiddenKindSlugs {
		if slices.Contains(slugs, slug) {
			return fmt.Errorf("forbidden kind %s", slug)
		}
	}

	if len(categorySeeds) < 10 {
		return fmt.Errorf("want at least 10 categories, got %d", len(categorySeeds))
	}

	catIDs := make(map[string]struct{}, len(categorySeeds))
	for _, c := range categorySeeds {
		if _, dup := catIDs[c.id]; dup {
			return fmt.Errorf("duplicate category id %s", c.id)
		}
		catIDs[c.id] = struct{}{}
		depth := categoryDepth(categorySeeds, c.id)
		if depth < 1 {
			return fmt.Errorf("category %s: cycle or missing parent", c.name)
		}
		if depth > 3 {
			return fmt.Errorf("category %s: depth %d > 3", c.name, depth)
		}
	}
	return nil
}

func checkAttrSchema(slug string, fields []attrField) error {
	seen := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		if f.Key == "" || f.Label == "" {
			return fmt.Errorf("kind %s: empty attr key/label", slug)
		}
		if _, dup := seen[f.Key]; dup {
			return fmt.Errorf("kind %s: duplicate attr %s", slug, f.Key)
		}
		seen[f.Key] = struct{}{}
		switch f.Type {
		case "string", "number", "boolean":
		default:
			return fmt.Errorf("kind %s: bad attr type %s", slug, f.Type)
		}
	}
	return nil
}
